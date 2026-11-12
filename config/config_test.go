package config_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jani123/dome/config"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type reader struct{ mock.Mock }

func (r *reader) ReadFile(p string) ([]byte, error) {
	a := r.Called(p)
	return a.Get(0).([]byte), a.Error(1)
}

type decrypter struct{ mock.Mock }

func (d *decrypter) Decrypt(ctx context.Context, b []byte) ([]byte, error) {
	a := d.Called(b)
	return a.Get(0).([]byte), a.Error(1)
}

type (
	cfg struct {
		Name string `yaml:"name"`
	}
	secret struct {
		Token string `yaml:"token"`
	}
)

func TestTypedLocalAndSecrets(t *testing.T) {
	r := &reader{}
	r.On("ReadFile", "config").Return([]byte("name: hello"), nil)
	r.On("ReadFile", "secrets").Return([]byte("encrypted"), nil)
	doc, err := (config.Local{Path: "config", SecretsPath: "secrets", Reader: r}).Load(context.Background())
	require.NoError(t, err)
	cleartext := []byte("token: sensitive")
	dec := &decrypter{}
	dec.On("Decrypt", []byte("encrypted")).Return(cleartext, nil)
	c, s, err := config.Parse[cfg, secret](context.Background(), doc, dec)
	require.NoError(t, err)
	require.Equal(t, "hello", c.Name)
	require.Equal(t, "sensitive", s.Token)
	require.Equal(t, make([]byte, len(cleartext)), cleartext)
	r.AssertExpectations(t)
	dec.AssertExpectations(t)
}

func TestYAMLValidationAndRedaction(t *testing.T) {
	for _, b := range []string{"unknown: sensitive", "name: [sensitive]", "name: hi\n---\nname: second"} {
		_, _, err := config.Parse[cfg, secret](context.Background(), config.Document{Configuration: []byte(b)}, nil)
		require.Error(t, err)
		require.NotContains(t, err.Error(), "sensitive")
	}
	r := &reader{}
	cause := errors.New("sensitive path")
	r.On("ReadFile", "config").Return([]byte(nil), cause)
	_, err := (config.Local{Path: "config", Reader: r}).Load(context.Background())
	require.ErrorIs(t, err, cause)
	require.NotContains(t, err.Error(), "sensitive")
}
