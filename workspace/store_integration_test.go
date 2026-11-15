//go:build integration

package workspace_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/jani123/dome/workspace"
	"github.com/stretchr/testify/require"
)

func TestRealWorkspacePublication(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	s := workspace.OSStore{}
	p := workspace.Profile{SchemaVersion: 1, Source: "git", GitURL: "https://example.com/config", GitRef: "main"}.Defaults()
	require.NoError(t, s.Save(ctx, dir, p))
	profile := filepath.Join(dir, ".dome/agent.yaml")
	info, err := os.Stat(profile)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	info, err = os.Stat(filepath.Dir(profile))
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o700), info.Mode().Perm())
	got, found, err := s.Load(ctx, dir, "")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, p, got)
	before, err := os.ReadFile(profile)
	require.NoError(t, err)
	require.NoError(t, s.Save(ctx, dir, p))
	p.GitRef = "other"
	require.Error(t, s.Save(ctx, dir, p))
	after, err := os.ReadFile(profile)
	require.NoError(t, err)
	require.Equal(t, before, after)
	_, err = os.Stat(filepath.Join(dir, ".dome/state.db"))
	require.True(t, os.IsNotExist(err))
	entries, err := os.ReadDir(filepath.Join(dir, ".dome"))
	require.NoError(t, err)
	require.Len(t, entries, 1)
}

func TestWorkspaceRejectsSymlinksAndCancellation(t *testing.T) {
	for _, name := range []string{"ancestor", "profile"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			target := t.TempDir()
			s := workspace.OSStore{}
			p := workspace.Profile{SchemaVersion: 1, Source: "local", Config: "config.yaml"}.Defaults()
			if name == "ancestor" {
				require.NoError(t, os.Symlink(target, filepath.Join(dir, ".dome")))
			} else {
				require.NoError(t, os.Mkdir(filepath.Join(dir, ".dome"), 0o700))
				require.NoError(t, os.WriteFile(filepath.Join(target, "file"), []byte("preserve"), 0o600))
				require.NoError(t, os.Symlink(filepath.Join(target, "file"), filepath.Join(dir, ".dome/agent.yaml")))
			}
			require.Error(t, s.Save(context.Background(), dir, p))
			_, _, err := s.Load(context.Background(), dir, "")
			require.Error(t, err)
			if name == "profile" {
				data, err := os.ReadFile(filepath.Join(target, "file"))
				require.NoError(t, err)
				require.Equal(t, "preserve", string(data))
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	dir := t.TempDir()
	require.ErrorIs(t, (workspace.OSStore{}).Save(ctx, dir, workspace.Profile{}), context.Canceled)
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Empty(t, entries)
}

func TestConcurrentWorkspaceInitialization(t *testing.T) {
	dir := t.TempDir()
	s := workspace.OSStore{}
	var wg sync.WaitGroup
	results := make(chan error, 8)
	for i := range 8 {
		wg.Go(func() {
			p := workspace.Profile{SchemaVersion: 1, Source: "git", GitURL: "https://example.com/config", GitRef: string(rune('a' + i))}.Defaults()
			results <- s.Save(context.Background(), dir, p)
		})
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		}
	}
	require.Equal(t, 1, success)
	got, found, err := s.Load(context.Background(), dir, "")
	require.NoError(t, err)
	require.True(t, found)
	require.NoError(t, got.Validate())
	entries, err := os.ReadDir(filepath.Join(dir, ".dome"))
	require.NoError(t, err)
	require.Len(t, entries, 1)
}

func TestWorkspaceReferencesAndExplicitProfiles(t *testing.T) {
	dir := t.TempDir()
	s := workspace.OSStore{}
	require.Error(t, s.CheckFile(context.Background(), dir))
	path := filepath.Join(dir, "settings.yaml")
	require.NoError(t, os.WriteFile(path, []byte("schema_version: 1\nsource: local\nconfig: config.yaml\n"), 0o600))
	require.NoError(t, s.CheckFile(context.Background(), path))
	p, found, err := s.Load(context.Background(), dir, "settings.yaml")
	require.NoError(t, err)
	require.True(t, found)
	require.NoError(t, p.Defaults().Validate())
}

func TestDocumentedRuntimeProfile(t *testing.T) {
	data, err := os.ReadFile("../docs/content/configuration.md")
	require.NoError(t, err)
	body := string(data)
	start := strings.Index(body, "```yaml\n")
	require.NotEqual(t, -1, start)
	body = body[start+len("```yaml\n"):]
	end := strings.Index(body, "```")
	require.NotEqual(t, -1, end)
	p, err := workspace.Decode([]byte(body[:end]))
	require.NoError(t, err)
	require.NoError(t, p.Defaults().Validate())
}

func TestOversizedSaveFailsBeforePublication(t *testing.T) {
	dir := t.TempDir()
	p := workspace.Profile{SchemaVersion: 1, Source: "git", GitURL: "https://example.com/" + strings.Repeat("x", 64<<10)}.Defaults()
	require.NoError(t, p.Validate())
	require.Error(t, (workspace.OSStore{}).Save(context.Background(), dir, p))
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Empty(t, entries)
}

func TestFilesystemCausesRemainInspectable(t *testing.T) {
	err := (workspace.OSStore{}).CheckFile(context.Background(), filepath.Join(t.TempDir(), "private-secret-name"))
	require.ErrorIs(t, err, os.ErrNotExist)
	require.NotContains(t, err.Error(), "private-secret-name")
}
