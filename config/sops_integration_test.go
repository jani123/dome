//go:build integration

package config_test

import (
	"context"
	"testing"

	"github.com/jani123/dome/config"
	"github.com/jani123/dome/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestRealSOPSHostKeyAndWrongKey(t *testing.T) {
	encrypted, key, err := testutil.EncryptYAML([]byte("token: do-not-log-this"))
	require.NoError(t, err)
	t.Setenv("SOPS_AGE_KEY", key)
	_, s, err := config.Parse[cfg, secret](context.Background(), config.Document{Secrets: encrypted}, nil)
	require.NoError(t, err)
	require.Equal(t, "do-not-log-this", s.Token)
	_, wrong, err := testutil.EncryptYAML([]byte("token: wrong"))
	require.NoError(t, err)
	t.Setenv("SOPS_AGE_KEY", wrong)
	t.Setenv("SOPS_AGE_KEY_FILE", "/nonexistent/dome-test-key")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	_, _, err = config.Parse[cfg, secret](context.Background(), config.Document{Secrets: encrypted}, nil)
	require.Error(t, err)
	require.NotContains(t, err.Error(), "do-not-log-this")
	require.NotContains(t, err.Error(), key)
}
