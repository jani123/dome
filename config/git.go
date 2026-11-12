package config

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"

	git "github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"golang.org/x/crypto/ssh/knownhosts"
)

type (
	// GitOptions identifies committed configuration; SSH requires a key and known_hosts.
	GitOptions struct {
		URL         string
		Ref         string
		Path        string
		SecretsPath string
		SSHKey      string
		KnownHosts  string
		AllowHTTP   bool
	}
	// GitTransport is the injectable committed-blob retrieval boundary.
	GitTransport interface {
		Fetch(context.Context, GitOptions) (Document, error)
	}
	// Git loads typed configuration and encrypted secrets from one commit.
	Git struct {
		Options   GitOptions
		Transport GitTransport
	}
)

// Load validates the source and fetches committed configuration through its transport.
func (s Git) Load(ctx context.Context) (Document, error) {
	if err := s.Options.Validate(); err != nil {
		return Document{}, hidden(err, "invalid Git configuration source")
	}
	client := s.Transport
	if client == nil {
		client = GoGit{}
	}
	d, err := client.Fetch(ctx, s.Options)
	if err != nil {
		return Document{}, hidden(err, "Git configuration retrieval failed")
	}
	return d, nil
}

// Validate rejects unsafe paths, unsupported transports, and missing SSH verification.
func (o GitOptions) Validate() error {
	u, err := url.Parse(o.URL)
	if err != nil || u.Host == "" || (u.Scheme != "ssh" && u.Scheme != "https" && !(u.Scheme == "http" && o.AllowHTTP)) {
		return fmt.Errorf("Git URL must use ssh:// or https://; unencrypted HTTP requires AllowHTTP")
	}
	if u.User != nil {
		if _, hasPassword := u.User.Password(); hasPassword {
			return fmt.Errorf("credentials in URLs are unsupported")
		}
	}
	if o.Path == "" {
		return fmt.Errorf("configuration path is required")
	}
	for _, p := range []string{o.Path, o.SecretsPath} {
		if p != "" && (path.IsAbs(p) || path.Clean(p) != p || p == ".." || strings.HasPrefix(p, "../")) {
			return fmt.Errorf("Git paths must be clean repository-relative paths")
		}
	}
	if u.Scheme == "ssh" && (o.SSHKey == "" || o.KnownHosts == "") {
		return fmt.Errorf("SSH requires a private key and known_hosts file")
	}
	return nil
}

// GoGit reads committed blobs directly into memory; secrets are never checked out.
type GoGit struct{}

// Fetch retrieves committed blobs in memory with SSH host verification when applicable.
func (GoGit) Fetch(ctx context.Context, o GitOptions) (Document, error) {
	var auth transport.AuthMethod
	u, err := url.Parse(o.URL)
	if err != nil {
		return Document{}, err
	}
	if u.Scheme == "ssh" {
		user := "git"
		if u.User != nil && u.User.Username() != "" {
			user = u.User.Username()
		}
		key, err := gitssh.NewPublicKeysFromFile(user, o.SSHKey, "")
		if err != nil {
			return Document{}, err
		}
		callback, err := knownhosts.New(o.KnownHosts)
		if err != nil {
			return Document{}, err
		}
		key.HostKeyCallback = callback
		auth = key
	}
	repo, err := git.Init(memory.NewStorage(), nil)
	if err != nil {
		return Document{}, err
	}
	remote, err := repo.CreateRemote(&gitconfig.RemoteConfig{Name: "origin", URLs: []string{o.URL}, Fetch: []gitconfig.RefSpec{"+refs/heads/*:refs/heads/*", "+refs/tags/*:refs/tags/*"}})
	if err != nil {
		return Document{}, err
	}
	if err = repo.FetchContext(ctx, &git.FetchOptions{Auth: auth, Tags: git.AllTags}); err != nil {
		return Document{}, err
	}
	ref := o.Ref
	if ref == "" {
		ref = "HEAD"
	}
	// Fetching explicit branches must not depend on the remote's default HEAD,
	// which may still refer to an unborn branch in a newly created repository.
	if ref == "HEAD" {
		refs, err := remote.ListContext(ctx, &git.ListOptions{Auth: auth})
		if err != nil {
			return Document{}, err
		}
		found := false
		for _, r := range refs {
			if r.Name() == plumbing.HEAD {
				if err = repo.Storer.SetReference(r); err != nil {
					return Document{}, err
				}
				found = true
				break
			}
		}
		if !found {
			return Document{}, plumbing.ErrReferenceNotFound
		}
	}

	hash, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil && !strings.HasPrefix(ref, "refs/") {
		hash, err = repo.ResolveRevision(plumbing.Revision("refs/remotes/origin/" + ref))
	}
	if err != nil {
		return Document{}, err
	}
	commit, err := repo.CommitObject(*hash)
	if err != nil {
		return Document{}, err
	}
	tree, err := commit.Tree()
	if err != nil {
		return Document{}, err
	}
	read := func(p string) ([]byte, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		f, err := tree.File(p)
		if err != nil {
			return nil, err
		}
		if f.Size > 16<<20 {
			return nil, fmt.Errorf("configuration blob exceeds 16 MiB")
		}
		reader, err := f.Reader()
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		return io.ReadAll(reader)
	}
	b, err := read(o.Path)
	if err != nil {
		return Document{}, err
	}
	d := Document{Configuration: b, Revision: hash.String()}
	if o.SecretsPath != "" {
		d.Secrets, err = read(o.SecretsPath)
		if err != nil {
			return Document{}, err
		}
	}
	return d, nil
}
