// Package workspace stores non-secret runtime settings separately from resource YAML.
package workspace

import (
	"bytes"
	"errors"
	"io"
	"net/url"
	"path/filepath"

	"github.com/jani123/dome/config"
	"github.com/jani123/dome/core/fault"
	"gopkg.in/yaml.v3"
)

// Profile is the versioned host bootstrap configuration. Credential fields are
// file references, never key material. Relative host paths use the workspace.
type Profile struct {
	SchemaVersion int    `yaml:"schema_version"`
	Source        string `yaml:"source"`
	State         string `yaml:"state"`
	Config        string `yaml:"config,omitempty"`
	Secrets       string `yaml:"secrets,omitempty"`
	GitURL        string `yaml:"git_url,omitempty"`
	GitRef        string `yaml:"git_ref,omitempty"`
	GitPath       string `yaml:"git_path,omitempty"`
	SSHKey        string `yaml:"ssh_key,omitempty"`
	KnownHosts    string `yaml:"known_hosts,omitempty"`
	AllowHTTP     bool   `yaml:"git_allow_http,omitempty"`
}

// Decode strictly reads one bounded YAML document without echoing its contents.
func Decode(data []byte) (Profile, error) {
	var p Profile
	if len(data) > 64<<10 {
		return p, invalid("workspace profile exceeds 64 KiB")
	}
	d := yaml.NewDecoder(bytes.NewReader(data))
	d.KnownFields(true)
	if err := d.Decode(&p); err != nil {
		return Profile{}, concealed(err, "invalid workspace profile YAML")
	}
	var extra any
	if err := d.Decode(&extra); !errors.Is(err, io.EOF) {
		return Profile{}, invalid("workspace profile must contain one document")
	}
	return p, nil
}

func invalid(message string) error { return fault.New(fault.Configuration, message) }

// Defaults fills optional schema values. Required source inputs are never inferred.
func (p Profile) Defaults() Profile {
	if p.State == "" {
		p.State = ".dome/state.db"
	}
	if p.Source == "git" {
		if p.GitRef == "" {
			p.GitRef = "HEAD"
		}
		if p.GitPath == "" {
			p.GitPath = "config.yaml"
		}
	}
	return p
}

// Validate checks bootstrap settings without loading resource data or secrets.
func (p Profile) Validate() error {
	if p.SchemaVersion != 1 {
		return invalid("unsupported workspace profile schema version")
	}
	if p.State == "" {
		return invalid("workspace state path is required")
	}
	switch p.Source {
	case "local":
		if p.Config == "" {
			return invalid("local configuration path is required")
		}
		if p.GitURL != "" || p.GitRef != "" || p.GitPath != "" || p.SSHKey != "" || p.KnownHosts != "" || p.AllowHTTP {
			return invalid("local source cannot include Git settings")
		}
	case "git":
		u, err := url.Parse(p.GitURL)
		if err != nil || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "ssh" && u.User != nil) {
			return invalid("Git workspace URL must not include HTTP credentials, query parameters or fragments")
		}
		if p.Config != "" {
			return invalid("Git source cannot include local configuration")
		}
		if p.GitRef == "" {
			return invalid("Git ref is required")
		}
		if err := p.GitOptions().Validate(); err != nil {
			return invalid("invalid Git workspace settings; check URL, repository paths and SSH file references")
		}
	default:
		return invalid("workspace source must be local or git")
	}
	return nil
}

// GitOptions returns transport settings; caller resolves host paths first.
func (p Profile) GitOptions() config.GitOptions {
	return config.GitOptions{URL: p.GitURL, Ref: p.GitRef, Path: p.GitPath, SSHKey: p.SSHKey, KnownHosts: p.KnownHosts, AllowHTTP: p.AllowHTTP, SecretsPath: p.Secrets}
}

// Resolve anchors host file references without changing process working directory.
func (p Profile) Resolve(dir string) Profile {
	host := func(s string) string {
		if s == "" || filepath.IsAbs(s) {
			return s
		}
		return filepath.Join(dir, s)
	}
	p.State = host(p.State)
	p.Config = host(p.Config)
	p.SSHKey = host(p.SSHKey)
	p.KnownHosts = host(p.KnownHosts)
	if p.Source == "local" {
		p.Secrets = host(p.Secrets)
	}
	return p
}
