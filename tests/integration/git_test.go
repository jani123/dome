//go:build integration

package integration_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	gittransport "github.com/go-git/go-git/v5/plumbing/transport"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	domeconfig "github.com/jani123/dome/config"
	"github.com/jani123/dome/core/reconcile"
	"github.com/jani123/dome/core/resource"
	"github.com/jani123/dome/internal/testutil"
	"github.com/moby/moby/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	tc "github.com/testcontainers/testcontainers-go"
	tcnetwork "github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func TestGitIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dome real Git HTTP and SSH integration")
}

var (
	binary string
	_      = BeforeSuite(func(ctx SpecContext) {
		// Published test ports are local Docker endpoints, not internet egress.
		// Configure this before any HTTP Git client initializes its proxy settings.
		provider, err := tc.NewDockerProvider()
		Expect(err).NotTo(HaveOccurred())
		defer provider.Close()
		host, err := provider.DaemonHost(ctx)
		Expect(err).NotTo(HaveOccurred())
		for _, name := range []string{"NO_PROXY", "no_proxy"} {
			value := os.Getenv(name)
			if value != "" {
				value += ","
			}
			GinkgoT().Setenv(name, value+host)
		}
		binary, err = testutil.BuildAgent(ctx, GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
	}, NodeTimeout(3*time.Minute))
	_ = Describe("Git-backed configuration", func() {
		DescribeTable("retrieves revisions, reconciles updates, persists state, and handles failures", func(ctx SpecContext, transportMode string) {
			network, err := tcnetwork.New(ctx)
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { Expect(network.Remove(context.Background())).To(Succeed()) })
			_, private, err := ed25519.GenerateKey(rand.Reader)
			Expect(err).NotTo(HaveOccurred())
			block, err := ssh.MarshalPrivateKey(private, "dome integration")
			Expect(err).NotTo(HaveOccurred())
			privatePEM := pem.EncodeToMemory(block)
			signer, err := ssh.NewSignerFromKey(private)
			Expect(err).NotTo(HaveOccurred())
			public := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(signer.PublicKey())))
			buildArgs := map[string]*string{}
			for _, name := range []string{"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY", "http_proxy", "https_proxy", "no_proxy"} {
				if value := os.Getenv(name); value != "" {
					buildArgs[name] = &value
				}
			}
			request := tc.ContainerRequest{
				FromDockerfile: tc.FromDockerfile{
					Context: filepath.Join(testutil.Root(), "tests/integration/gitserver"), Repo: "dome-git-test", Tag: "local", KeepImage: true, BuildArgs: buildArgs, BuildLogWriter: GinkgoWriter,
					// The devcontainer's proxy is reachable from the daemon's host network.
					// Only the image build uses this network; both runtime containers are isolated.
					BuildOptionsModifier: func(options *client.ImageBuildOptions) { options.NetworkMode = "host" },
				},
				Env: map[string]string{"TEST_GIT_AUTHORIZED_KEY": public}, Networks: []string{network.Name}, NetworkAliases: map[string][]string{network.Name: {"gitserver"}}, ExposedPorts: []string{"2222/tcp", "8080/tcp"}, WaitingFor: wait.ForAll(wait.ForListeningPort("2222/tcp"), wait.ForListeningPort("8080/tcp")).WithStartupTimeout(time.Minute),
			}

			server, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{ContainerRequest: request, Started: true})
			if server != nil {
				DeferCleanup(func() {
					if CurrentSpecReport().Failed() {
						logs, err := server.Logs(context.Background())
						if err == nil {
							_, _ = io.Copy(GinkgoWriter, logs)
							_ = logs.Close()
						}
					}
					Expect(server.Terminate(context.Background(), tc.StopTimeout(time.Second))).To(Succeed())
				})
			}
			Expect(err).NotTo(HaveOccurred())
			host, err := server.Host(ctx)
			Expect(err).NotTo(HaveOccurred())
			port, err := server.MappedPort(ctx, "2222/tcp")
			Expect(err).NotTo(HaveOccurred())
			address := net.JoinHostPort(host, port.Port())
			publicReader, err := server.CopyFileFromContainer(ctx, "/run/git-test-server/ssh_host_ed25519_key.pub")
			Expect(err).NotTo(HaveOccurred())
			hostPublic, err := io.ReadAll(publicReader)
			Expect(err).NotTo(HaveOccurred())
			Expect(publicReader.Close()).To(Succeed())
			dir := GinkgoT().TempDir()
			keyPath := filepath.Join(dir, "key")
			hostsPath := filepath.Join(dir, "known_hosts")
			Expect(os.WriteFile(keyPath, privatePEM, 0o600)).To(Succeed())
			knownLine := fmt.Sprintf("[%s]:%s,[gitserver]:2222 %s\n", host, port.Port(), strings.TrimSpace(string(hostPublic)))
			Expect(os.WriteFile(hostsPath, []byte(knownLine), 0o600)).To(Succeed())
			callback, err := knownhosts.New(hostsPath)
			Expect(err).NotTo(HaveOccurred())
			sshClient, err := ssh.Dial("tcp", address, &ssh.ClientConfig{User: "git", Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)}, HostKeyCallback: callback, Timeout: 10 * time.Second})
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { Expect(sshClient.Close()).To(Succeed()) })
			admin := func(command string) string {
				session, err := sshClient.NewSession()
				Expect(err).NotTo(HaveOccurred())
				defer session.Close()
				out, err := session.CombinedOutput(command)
				Expect(err).NotTo(HaveOccurred(), string(out))
				return string(out)
			}
			admin("git init --bare --initial-branch=unborn /srv/git/configuration.git")
			admin("git --git-dir=/srv/git/configuration.git config http.receivepack true")
			Expect(admin("git --git-dir=/srv/git/configuration.git rev-parse --is-bare-repository")).To(ContainSubstring("true"))
			repositoryURL := "ssh://git@" + address + "/srv/git/configuration.git"
			if transportMode == "http" {
				httpPort, err := server.MappedPort(ctx, "8080/tcp")
				Expect(err).NotTo(HaveOccurred())
				repositoryURL = "http://" + net.JoinHostPort(host, httpPort.Port()) + "/git/configuration.git"
			}

			writerPath := filepath.Join(dir, "writer")
			repo, err := git.PlainInit(writerPath, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(repo.Storer.SetReference(plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.NewBranchReferenceName("main")))).To(Succeed())
			_, err = repo.CreateRemote(&config.RemoteConfig{Name: "origin", URLs: []string{repositoryURL}})
			Expect(err).NotTo(HaveOccurred())
			sshAuth, err := gitssh.NewPublicKeysFromFile("git", keyPath, "")
			Expect(err).NotTo(HaveOccurred())
			sshAuth.HostKeyCallback = callback
			var auth gittransport.AuthMethod = sshAuth
			if transportMode == "http" {
				auth = nil
			}
			sequence := 0
			push := func(files map[string][]byte) string {
				worktree, err := repo.Worktree()
				Expect(err).NotTo(HaveOccurred())
				for name, body := range files {
					Expect(os.WriteFile(filepath.Join(writerPath, name), body, 0o600)).To(Succeed())
					_, err = worktree.Add(name)
					Expect(err).NotTo(HaveOccurred())
				}
				sequence++
				hash, err := worktree.Commit(fmt.Sprintf("configuration %d", sequence), &git.CommitOptions{Author: &object.Signature{Name: "Dome test", Email: "test@example.invalid", When: time.Now()}})
				Expect(err).NotTo(HaveOccurred())
				Expect(repo.PushContext(ctx, &git.PushOptions{Auth: auth, Progress: GinkgoWriter, RefSpecs: []config.RefSpec{"refs/heads/main:refs/heads/main"}})).To(Succeed())
				return hash.String()
			}
			first := push(map[string][]byte{"config.yaml": []byte("content: first\nversion: '1'\n")})
			fetched, fetchErr := (domeconfig.GoGit{}).Fetch(ctx, domeconfig.GitOptions{URL: repositoryURL, AllowHTTP: transportMode == "http", Ref: "main", Path: "config.yaml", SSHKey: keyPath, KnownHosts: hostsPath})
			Expect(fetchErr).NotTo(HaveOccurred())
			Expect(fetched.Revision).To(Equal(first))
			// Verify explicit commits, annotated tags, and the server's default HEAD.
			_, err = repo.CreateTag("v1", plumbing.NewHash(first), &git.CreateTagOptions{Tagger: &object.Signature{Name: "Dome test", Email: "test@example.invalid", When: time.Now()}, Message: "first configuration"})
			Expect(err).NotTo(HaveOccurred())
			Expect(repo.PushContext(ctx, &git.PushOptions{Auth: auth, RefSpecs: []config.RefSpec{"refs/tags/v1:refs/tags/v1"}})).To(Succeed())
			admin("git --git-dir=/srv/git/configuration.git symbolic-ref HEAD refs/heads/main")
			for _, ref := range []string{"", "HEAD", first, "v1", "refs/heads/main"} {
				doc, err := (domeconfig.GoGit{}).Fetch(ctx, domeconfig.GitOptions{URL: repositoryURL, AllowHTTP: transportMode == "http", Ref: ref, Path: "config.yaml", SSHKey: keyPath, KnownHosts: hostsPath})
				Expect(err).NotTo(HaveOccurred(), ref)
				Expect(doc.Revision).To(Equal(first), ref)
			}

			agentContainer, err := testutil.StartAgent(ctx, binary, []string{network.Name})
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { Expect(agentContainer.Cleanup(context.Background())).To(Succeed()) })
			Expect(agentContainer.CopyToContainer(ctx, privatePEM, "/data/git.key", 0o600)).To(Succeed())
			Expect(agentContainer.CopyToContainer(ctx, []byte(knownLine), "/data/known_hosts", 0o600)).To(Succeed())
			args := []string{"apply", "--git-ref", "main"}
			if transportMode == "http" {
				args = append(args, "--git-url", "http://gitserver:8080/git/configuration.git", "--git-allow-http")
			} else {
				args = append(args, "--git-url", "ssh://git@gitserver:2222/srv/git/configuration.git", "--ssh-key", "/data/git.key", "--known-hosts", "/data/known_hosts")
			}
			apply := func(extra ...string) (int, string) {
				code, out, err := agentContainer.Agent(ctx, append(append([]string{}, args...), extra...)...)
				Expect(err).NotTo(HaveOccurred())
				return code, out
			}
			read := func(path string) string {
				code, out, err := testutil.Exec(ctx, agentContainer.Container, "cat", path)
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(0), out)
				return out
			}
			status := func() string {
				code, out, err := agentContainer.Agent(ctx, "status")
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(0), out)
				return out
			}
			code, out := apply()
			Expect(code).To(Equal(0), out)
			Expect(read("/data/managed.txt")).To(Equal("first"))
			Expect(read("/data/version.txt")).To(Equal("1"))
			Expect(status()).To(ContainSubstring(first))
			second := push(map[string][]byte{"config.yaml": []byte("content: second\nversion: '2'\n")})
			code, out = apply()
			Expect(code).To(Equal(0), out)
			Expect(read("/data/managed.txt")).To(Equal("second"))
			Expect(read("/data/version.txt")).To(Equal("2"))
			Expect(status()).To(ContainSubstring(second))
			code, out = apply()
			Expect(code).To(Equal(0), out)
			var p reconcile.Plan
			Expect(json.Unmarshal([]byte(out), &p)).To(Succeed())
			for _, change := range p.Changes {
				Expect(change.Action).To(Equal(resource.Noop))
			}
			// Restart the process/container while retaining its isolated state volume.
			timeout := time.Second
			Expect(agentContainer.Stop(ctx, &timeout)).To(Succeed())
			Expect(agentContainer.Start(ctx)).To(Succeed())
			Expect(status()).To(ContainSubstring(second))
			// Unknown typed fields fail before pruning; the old configuration stays applied.
			invalid := push(map[string][]byte{"config.yaml": []byte("managed: false\nunknown: invalid\n")})
			code, out = apply()
			Expect(code).To(Equal(2), out)
			Expect(status()).To(ContainSubstring(second))
			Expect(status()).NotTo(ContainSubstring(invalid))
			Expect(read("/data/managed.txt")).To(Equal("second"))
			// Failed resource transitions leave the last successful revision unchanged.
			failed := push(map[string][]byte{"config.yaml": []byte("content: second\nversion: '3'\nfail_transition: true\n")})
			code, out = apply()
			Expect(code).NotTo(Equal(0), out)
			Expect(status()).To(ContainSubstring(second))
			Expect(status()).NotTo(ContainSubstring(failed))
			Expect(read("/data/version.txt")).To(Equal("2"))
			// Git-hosted encrypted secrets use a host-specific key inside the agent container.
			encrypted, hostKey, err := testutil.EncryptYAML([]byte("token: integration-secret"))
			Expect(err).NotTo(HaveOccurred())
			Expect(agentContainer.CopyToContainer(ctx, []byte(hostKey), "/data/age.key", 0o600)).To(Succeed())
			secretCommit := push(map[string][]byte{"config.yaml": []byte("content: with-secrets\nversion: '3'\n"), "secrets.yaml": encrypted})
			secretArgs := append(append([]string{}, args...), "--secrets", "secrets.yaml")
			envArgs := append([]string{"env", "SOPS_AGE_KEY_FILE=/data/age.key", "agent", "--state", "/data/state.db", "--json"}, secretArgs...)
			code, out, err = testutil.Exec(ctx, agentContainer.Container, envArgs...)
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(0), out)
			Expect(status()).To(ContainSubstring(secretCommit))
			Expect(read("/data/managed.txt")).To(Equal("with-secrets"))
			Expect(out).NotTo(ContainSubstring("integration-secret"))
			_, wrongKey, err := testutil.EncryptYAML([]byte("token: wrong"))
			Expect(err).NotTo(HaveOccurred())
			Expect(agentContainer.CopyToContainer(ctx, []byte(wrongKey), "/data/age.key", 0o600)).To(Succeed())
			push(map[string][]byte{"config.yaml": []byte("managed: false\nversion: '3'\n")})
			code, out, err = testutil.Exec(ctx, agentContainer.Container, envArgs...)
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(2), out)
			Expect(status()).To(ContainSubstring(secretCommit))
			Expect(read("/data/managed.txt")).To(Equal("with-secrets"))
			if transportMode == "ssh" {

				// Host verification and missing refs must fail closed before reconciliation.
				Expect(agentContainer.CopyToContainer(ctx, []byte(""), "/data/known_hosts", 0o600)).To(Succeed())
				code, out = apply()
				Expect(code).To(Equal(2), out)
				Expect(status()).To(ContainSubstring(secretCommit))
				Expect(agentContainer.CopyToContainer(ctx, []byte(knownLine), "/data/known_hosts", 0o600)).To(Succeed())
				_, wrongPrivate, err := ed25519.GenerateKey(rand.Reader)
				Expect(err).NotTo(HaveOccurred())
				wrongBlock, err := ssh.MarshalPrivateKey(wrongPrivate, "unauthorized test")
				Expect(err).NotTo(HaveOccurred())
				Expect(agentContainer.CopyToContainer(ctx, pem.EncodeToMemory(wrongBlock), "/data/unauthorized.key", 0o600)).To(Succeed())
				code, out = apply("--ssh-key", "/data/unauthorized.key")
				Expect(code).To(Equal(2), out)
				Expect(status()).To(ContainSubstring(secretCommit))
			} else {
				code, out = apply("--git-allow-http=false")
				Expect(code).To(Equal(2), out)
				Expect(status()).To(ContainSubstring(secretCommit))
			}

			code, out = apply("--git-ref", "does-not-exist")
			Expect(code).To(Equal(2), out)
			Expect(status()).To(ContainSubstring(secretCommit))
			Expect(read("/data/managed.txt")).To(Equal("with-secrets"))
		}, Entry("SSH with verified host and client keys", "ssh", SpecTimeout(5*time.Minute)), Entry("Git smart HTTP with explicit plaintext opt-in", "http", SpecTimeout(5*time.Minute)))
	})
)
