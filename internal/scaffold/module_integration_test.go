//go:build integration

package scaffold

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestModulePublication(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	dest := filepath.Join(root, "module")
	require.NoError(t, os.Mkdir(dest, 0o755))
	files := map[string]File{"scripts/run.sh": {[]byte("#!/bin/sh\n"), 0o755}, "go.mod": {[]byte("module example.com/demo\n"), 0o644}}
	fail := errors.New("failed preparation")
	require.ErrorIs(t, PublishModule(ctx, dest, files, func(context.Context, string) error { return fail }), fail)
	entries, err := os.ReadDir(dest)
	require.NoError(t, err)
	require.Empty(t, entries)
	require.NoError(t, PublishModule(ctx, dest, files, nil))
	info, err := os.Stat(filepath.Join(dest, "scripts/run.sh"))
	require.NoError(t, err)
	require.EqualValues(t, 0o755, info.Mode().Perm())
	require.Error(t, PublishModule(ctx, dest, files, nil))
	empty := filepath.Join(root, "empty")
	require.NoError(t, os.Mkdir(empty, 0o755))
	require.NoError(t, os.Symlink(empty, filepath.Join(root, "link")))
	require.Error(t, PublishModule(ctx, filepath.Join(root, "link"), files, nil))
	for _, entry := range []string{".git", ".hidden"} {
		d := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(d, entry), nil, 0o600))
		require.Error(t, PublishModule(ctx, d, files, nil))
	}
	d := t.TempDir()
	cancelled, cancel := context.WithCancel(ctx)
	require.ErrorIs(t, PublishModule(cancelled, d, files, func(context.Context, string) error { cancel(); return nil }), context.Canceled)
	d = t.TempDir()
	require.Error(t, PublishModule(ctx, d, files, func(context.Context, string) error {
		return os.WriteFile(filepath.Join(d, "concurrent"), []byte("keep"), 0o600)
	}))
	b, err := os.ReadFile(filepath.Join(d, "concurrent"))
	require.NoError(t, err)
	require.Equal(t, "keep", string(b))
	require.Error(t, PublishModule(ctx, t.TempDir(), map[string]File{"../escape": {nil, 0o644}}, nil))
}

// TestGeneratedModule exercises actual commands in a standalone external module.
func TestGeneratedModule(t *testing.T) {
	if os.Getenv("DOME_MODULE_ACCEPTANCE") == "" { // No skip: the basic external compile always runs.
		t.Log("Set DOME_MODULE_ACCEPTANCE=1 for race, developer tools, docs and real container checks as well")
	}
	root, err := filepath.Abs("../..")
	require.NoError(t, err)
	dest := t.TempDir()
	files, err := RenderModule("example.com/module-fixture", "widget", "v0.1.0")
	require.NoError(t, err)
	prepare := func(ctx context.Context, stage string) error {
		cmd := exec.CommandContext(ctx, "go", "mod", "edit", "-replace=github.com/jani123/dome="+root)
		cmd.Dir = stage
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
		if os.Getenv("DOME_MODULE_ACCEPTANCE") != "" {
			return PrepareModule(ctx, stage)
		}
		return Tidy(ctx, stage)
	}
	require.NoError(t, PublishModule(context.Background(), dest, files, prepare))
	// Capture the fixture for optional broader verification outside this test.
	if path := os.Getenv("DOME_MODULE_FIXTURE_PATH"); path != "" {
		require.NoError(t, os.WriteFile(path, []byte(dest), 0o600))
	}
	commands := [][]string{
		{"go", "build", "./..."},
		{"go", "test", "./..."},
		{"go", "doc", "-all", "example.com/module-fixture"},
		{"go", "run", "scripts/check-knowledge.go", "-policy"},
		{"go", "run", "scripts/check-maintainer-skills.go"},
	}
	if os.Getenv("DOME_MODULE_ACCEPTANCE") != "" {
		// Exercise each documented check once. Full real container suites also
		// compile their tagged tests; compile-only probes add no coverage here.
		commands = [][]string{
			{"task", "dev:setup"},
			{"task", "build"},
			{"task", "check"},
			{"task", "test:race"},
			{"go", "doc", "-all", "example.com/module-fixture"},
			{"task", "check:artifacts"},
			{"task", "docs:build"},
			{"task", "check:workflow"},
			{"task", "test:integration"},
			{"task", "test:e2e"},
		}
	}
	for _, args := range commands {
		if !t.Run(strings.Join(args, " "), func(t *testing.T) {
			started := time.Now()
			cmd := exec.Command(args[0], args[1:]...)
			cmd.Dir = dest
			out, err := cmd.CombinedOutput()
			t.Logf("command completed in %s", time.Since(started).Round(time.Millisecond))
			require.NoError(t, err, "%v\n%s", args, out)
		}) {
			t.FailNow()
		}
	}
}

type publicationContext struct {
	context.Context
	calls int
	at    func(int)
}

func (c *publicationContext) Err() error { c.calls++; c.at(c.calls); return nil }
func TestModulePublicationRollbackOwnership(t *testing.T) {
	for _, modify := range []bool{false, true} {
		t.Run(fmt.Sprint(modify), func(t *testing.T) {
			dest := t.TempDir()
			files := map[string]File{"a/data": {[]byte("owned"), 0o644}, "b/run": {[]byte("owned"), 0o644}}
			ctx := &publicationContext{Context: context.Background(), at: func(call int) {
				if call == 3 {
					require.NoError(t, os.Mkdir(filepath.Join(dest, "b"), 0o755))
					if modify {
						require.NoError(t, os.WriteFile(filepath.Join(dest, "a/unrelated"), []byte("keep"), 0o644))
					}
				}
			}}
			require.Error(t, PublishModule(ctx, dest, files, nil))
			require.DirExists(t, filepath.Join(dest, "b"))
			if modify {
				b, err := os.ReadFile(filepath.Join(dest, "a/unrelated"))
				require.NoError(t, err)
				require.Equal(t, "keep", string(b))
			}
		})
	}
}
