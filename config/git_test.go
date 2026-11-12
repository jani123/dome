package config_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jani123/dome/config"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type transport struct{ mock.Mock }

func (tr *transport) Fetch(ctx context.Context, o config.GitOptions) (config.Document, error) {
	a := tr.Called(o)
	return a.Get(0).(config.Document), a.Error(1)
}

func TestGitSourceUsesResolvedRevision(t *testing.T) {
	o := config.GitOptions{URL: "ssh://git@example.invalid/config", Path: "config.yaml", SSHKey: "key", KnownHosts: "hosts", Ref: "main"}
	tr := &transport{}
	tr.On("Fetch", o).Return(config.Document{Configuration: []byte("name: test"), Revision: "commit"}, nil)
	d, err := (config.Git{Options: o, Transport: tr}).Load(context.Background())
	require.NoError(t, err)
	require.Equal(t, "commit", d.Revision)
	tr.AssertExpectations(t)
}

func TestGitValidationAndRedaction(t *testing.T) {
	for _, o := range []config.GitOptions{{URL: "http://example.invalid/x", Path: "config"}, {URL: "ssh://git@example.invalid/x", Path: "config"}, {URL: "https://example.invalid/x", Path: "../config"}, {URL: "https://user:secret@example.invalid/x", Path: "config"}, {URL: "https://example.invalid/x", Path: "/config"}} {
		require.Error(t, o.Validate())
	}
	o := config.GitOptions{URL: "https://example.invalid/x", Path: "config"}
	tr := &transport{}
	cause := errors.New("credential secret")
	tr.On("Fetch", o).Return(config.Document{}, cause)
	_, err := (config.Git{Options: o, Transport: tr}).Load(context.Background())
	require.ErrorIs(t, err, cause)
	require.NotContains(t, err.Error(), "secret")
}

func TestHTTPRequiresExplicitOptIn(t *testing.T) {
	o := config.GitOptions{URL: "http://git-test-server:8080/git/configuration.git", Path: "config.yaml"}
	require.Error(t, o.Validate())
	o.AllowHTTP = true
	require.NoError(t, o.Validate())
	tr := &transport{}
	tr.On("Fetch", o).Return(config.Document{Configuration: []byte("name: test"), Revision: "commit"}, nil)
	doc, err := (config.Git{Options: o, Transport: tr}).Load(context.Background())
	require.NoError(t, err)
	require.Equal(t, "commit", doc.Revision)
	tr.AssertExpectations(t)
}
