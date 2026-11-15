//go:build e2e

package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/jani123/dome/core/reconcile"
	"github.com/jani123/dome/core/resource"
	"github.com/jani123/dome/core/state"
	"github.com/jani123/dome/internal/testutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestE2E(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Dome container E2E") }

var (
	binary string
	_      = BeforeSuite(func(ctx SpecContext) {
		var err error
		binary, err = testutil.BuildAgent(ctx, GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
	}, NodeTimeout(3*time.Minute))
	_ = Describe("standalone agent", func() {
		var c *testutil.AgentContainer
		BeforeEach(func(ctx SpecContext) {
			var err error
			c, err = testutil.StartAgent(ctx, binary, nil)
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() {
				cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				Expect(c.Cleanup(cleanupCtx)).To(Succeed())
			})
		}, NodeTimeout(2*time.Minute))
		put := func(ctx context.Context, body string) {
			Expect(c.CopyToContainer(ctx, []byte(body), "/data/config.yaml", 0o600)).To(Succeed())
		}
		run := func(ctx context.Context, args ...string) string {
			code, out, err := c.Agent(ctx, args...)
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(0), out)
			return out
		}
		read := func(ctx context.Context, path string) string {
			code, out, err := testutil.Exec(ctx, c.Container, "cat", path)
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(0), out)
			return out
		}
		It("initializes host settings once and preserves history on repeated or failed setup", func(ctx SpecContext) {
			put(ctx, "content: initialized\n")
			invoke := func(args ...string) (int, string) {
				code, out, err := testutil.Exec(ctx, c.Container, append([]string{"agent", "--workspace", "/data", "--json"}, args...)...)
				Expect(err).NotTo(HaveOccurred())
				return code, out
			}
			code, out := invoke("init", "--config", "config.yaml", "--non-interactive")
			Expect(code).To(Equal(0), out)
			code, _, err := testutil.Exec(ctx, c.Container, "test", "!", "-e", "/data/.dome/state.db")
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(0))
			profile := read(ctx, "/data/.dome/agent.yaml")
			for _, args := range [][]string{{"plan"}, {"apply"}, {"apply"}} {
				code, out = invoke(args...)
				Expect(code).To(Equal(0), out)
			}
			Expect(out).To(ContainSubstring("no-op"))
			code, out, err = testutil.Exec(ctx, c.Container, "sh", "-c", "cd /tmp && agent --workspace /data status --json")
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(0), out)
			Expect(out).To(ContainSubstring("managed.txt"))
			Expect(read(ctx, "/data/managed.txt")).To(Equal("initialized"))
			beforeCode, before := invoke("operations", "list")
			Expect(beforeCode).To(Equal(0), before)
			code, out = invoke("init", "--non-interactive")
			Expect(code).To(Equal(0), out)
			code, out = invoke("init", "--git-url", "https://example.com/different", "--non-interactive")
			Expect(code).To(Equal(2), out)
			Expect(read(ctx, "/data/.dome/agent.yaml")).To(Equal(profile))
			code, out = invoke("operations", "list")
			Expect(code).To(Equal(0), out)
			Expect(out).To(Equal(before))
			code, out = invoke("status")
			Expect(code).To(Equal(0), out)
			Expect(out).To(ContainSubstring("managed.txt"))
			Expect(c.CopyToContainer(ctx, []byte("schema_version: 999\n"), "/data/.dome/agent.yaml", 0o600)).To(Succeed())
			code, out = invoke("apply")
			Expect(code).To(Equal(2), out)
			Expect(read(ctx, "/data/managed.txt")).To(Equal("initialized"))
		})
		It("creates, updates, detects no-op, prunes, and preserves unrelated files", func(ctx SpecContext) {
			Expect(c.CopyToContainer(ctx, []byte("keep"), "/data/unrelated", 0o600)).To(Succeed())
			put(ctx, "content: first\n")
			out := run(ctx, "plan", "--config", "/data/config.yaml")
			Expect(out).To(ContainSubstring(`"create"`))
			code, _, err := testutil.Exec(ctx, c.Container, "test", "!", "-e", "/data/managed.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(0))
			run(ctx, "apply", "--config", "/data/config.yaml")
			Expect(read(ctx, "/data/managed.txt")).To(Equal("first"))
			Expect(read(ctx, "/data/unrelated")).To(Equal("keep"))
			modeCode, mode, err := testutil.Exec(ctx, c.Container, "stat", "-c", "%a", "/data/managed.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(modeCode).To(Equal(0))
			Expect(mode).To(Equal("640\n"))
			put(ctx, "content: second\n")
			out = run(ctx, "apply", "--config", "/data/config.yaml")
			Expect(out).To(ContainSubstring(`"update"`))
			Expect(read(ctx, "/data/managed.txt")).To(Equal("second"))
			out = run(ctx, "apply", "--config", "/data/config.yaml")
			var p reconcile.Plan
			Expect(json.Unmarshal([]byte(out), &p)).To(Succeed())
			Expect(p.Changes).To(HaveLen(1))
			Expect(p.Changes[0].Action).To(Equal(resource.Noop))
			put(ctx, "managed: false\n")
			run(ctx, "apply", "--config", "/data/config.yaml")
			code, _, err = testutil.Exec(ctx, c.Container, "test", "!", "-e", "/data/managed.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(0))
			Expect(read(ctx, "/data/unrelated")).To(Equal("keep"))
		})
		It("executes parallel branches and their join", func(ctx SpecContext) {
			start := time.Now()
			run(ctx, "playbook", "run", "parallel-demo")
			Expect(time.Since(start)).To(BeNumerically("<", 2*time.Second))
			Expect(read(ctx, "/data/left.marker")).To(Equal("left\n"))
			Expect(read(ctx, "/data/right.marker")).To(Equal("right\n"))
			Expect(read(ctx, "/data/join.marker")).To(Equal("join\n"))
		})
		It("resumes a process that exits after its external write before SQLite completion", func(ctx SpecContext) {
			code, out, err := testutil.Exec(ctx, c.Container, "env", "DOME_CRASH_AFTER_WRITE=1", "agent", "--state", "/data/state.db", "playbook", "run", "crash-demo")
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(77), out)
			Expect(read(ctx, "/data/external.marker")).To(Equal("external-write\n"))
			var ops []state.Operation
			Expect(json.Unmarshal([]byte(run(ctx, "operations", "list")), &ops)).To(Succeed())
			Expect(ops).To(HaveLen(1))
			Expect(ops[0].Status).To(Equal("running"))
			id := ops[0].ID
			// If recovery blindly re-applied this step, the process would exit 77 again.
			code, out, err = testutil.Exec(ctx, c.Container, "env", "DOME_CRASH_AFTER_WRITE=1", "agent", "--state", "/data/state.db", "playbook", "run", "crash-demo", "--resume", id)
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(0), out)
			Expect(read(ctx, "/data/external.marker")).To(Equal("external-write\n"))
			Expect(run(ctx, "operations", "inspect", id)).To(ContainSubstring("succeeded"))
		})
		It("finishes an interrupted transition even when its resource already matches", func(ctx SpecContext) {
			put(ctx, "managed: false\nversion: '2'\n")
			args := []string{"env", "DOME_CRASH_AFTER_VERSION_WRITE=1", "agent", "--state", "/data/state.db", "apply", "--config", "/data/config.yaml"}
			code, out, err := testutil.Exec(ctx, c.Container, args...)
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(78), out)
			Expect(read(ctx, "/data/version.txt")).To(Equal("2"))
			var before struct {
				Resources []state.Managed `json:"resources"`
			}
			Expect(json.Unmarshal([]byte(run(ctx, "status")), &before)).To(Succeed())
			Expect(before.Resources).To(BeEmpty())
			code, out, err = testutil.Exec(ctx, c.Container, args...)
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(0), out)
			Expect(read(ctx, "/data/version.txt.finalized")).To(Equal("2"))
			Expect(run(ctx, "status")).To(ContainSubstring("demo-version"))
			var operations []state.Operation
			Expect(json.Unmarshal([]byte(run(ctx, "operations", "list")), &operations)).To(Succeed())
			executions := []state.Operation{}
			for _, operation := range operations {
				if operation.Action == "playbook" {
					executions = append(executions, operation)
				}
			}
			Expect(executions).To(HaveLen(1))
			Expect(executions[0].Status).To(Equal("succeeded"))
			Expect(executions[0].Nodes).To(HaveLen(2))
			for _, attempt := range executions[0].Attempts {
				Expect(attempt.Number).To(Equal(1))
			}
		})
		It("runs transition playbooks and preserves applied state on failure", func(ctx SpecContext) {
			put(ctx, "content: value\nversion: '1'\n")
			run(ctx, "apply", "--config", "/data/config.yaml")
			Expect(read(ctx, "/data/version.txt")).To(Equal("1"))
			put(ctx, "content: value\nversion: '2'\n")
			run(ctx, "apply", "--config", "/data/config.yaml")
			Expect(read(ctx, "/data/version.txt")).To(Equal("2"))
			before := run(ctx, "status")
			put(ctx, "content: value\nversion: '3'\nfail_transition: true\n")
			code, out, err := c.Agent(ctx, "apply", "--config", "/data/config.yaml")
			Expect(err).NotTo(HaveOccurred())
			Expect(code).NotTo(Equal(0), out)
			Expect(read(ctx, "/data/version.txt")).To(Equal("2"))
			Expect(run(ctx, "status")).To(Equal(before))
		})
		It("aborts invalid configuration without pruning and survives container restart", func(ctx SpecContext) {
			put(ctx, "content: keep\n")
			run(ctx, "apply", "--config", "/data/config.yaml")
			put(ctx, "unknown: invalid\n")
			code, out, err := c.Agent(ctx, "apply", "--config", "/data/config.yaml")
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(2), out)
			Expect(read(ctx, "/data/managed.txt")).To(Equal("keep"))
			timeout := time.Second
			Expect(c.Stop(ctx, &timeout)).To(Succeed())
			Expect(c.Start(ctx)).To(Succeed())
			Expect(run(ctx, "status")).To(ContainSubstring("managed.txt"))
			Expect(read(ctx, "/data/managed.txt")).To(Equal("keep"))
		})
		It("loads real SOPS secrets with a host key and rejects a wrong key", func(ctx SpecContext) {
			encrypted, key, err := testutil.EncryptYAML([]byte("token: test-secret-never-persist"))
			Expect(err).NotTo(HaveOccurred())
			Expect(c.CopyToContainer(ctx, encrypted, "/data/secrets.yaml", 0o600)).To(Succeed())
			Expect(c.CopyToContainer(ctx, []byte(key), "/data/host.key", 0o600)).To(Succeed())
			put(ctx, "content: value\n")
			code, out, err := testutil.Exec(ctx, c.Container, "env", "SOPS_AGE_KEY_FILE=/data/host.key", "agent", "--state", "/data/state.db", "apply", "--config", "/data/config.yaml", "--secrets", "/data/secrets.yaml")
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(0), out)
			Expect(out).NotTo(ContainSubstring("test-secret-never-persist"))
			_, wrong, err := testutil.EncryptYAML([]byte("token: wrong"))
			Expect(err).NotTo(HaveOccurred())
			Expect(c.CopyToContainer(ctx, []byte(wrong), "/data/host.key", 0o600)).To(Succeed())
			code, out, err = testutil.Exec(ctx, c.Container, "env", "SOPS_AGE_KEY_FILE=/data/host.key", "agent", "--state", "/data/state.db", "apply", "--config", "/data/config.yaml", "--secrets", "/data/secrets.yaml")
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(2), fmt.Sprint(out))
			Expect(out).NotTo(ContainSubstring("test-secret-never-persist"))
			Expect(read(ctx, "/data/managed.txt")).To(Equal("value"))
			code, out, err = testutil.Exec(ctx, c.Container, "sh", "-c", "if grep -a -q test-secret-never-persist /data/state.db*; then exit 1; fi")
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(0), out)
		})
	})
)
