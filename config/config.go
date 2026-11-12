// Package config loads user-defined typed YAML and SOPS secrets without an infrastructure DSL.
package config

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/getsops/sops/v3/decrypt"
	"github.com/jani123/dome/core/fault"
	"gopkg.in/yaml.v3"
)

// Document bytes are ephemeral. Never log or persist them.
type (
	// Document holds ephemeral configuration/secrets bytes and their source revision.
	Document struct {
		Configuration []byte
		Secrets       []byte
		Revision      string
	}
	// Source retrieves a complete configuration revision or returns an error.
	Source interface {
		Load(context.Context) (Document, error)
	}
	// Reader is the injectable local configuration filesystem boundary.
	Reader interface{ ReadFile(string) ([]byte, error) }
	// OSReader reads configuration from the host filesystem.
	OSReader struct{}
)

// ReadFile reads configuration bytes from the host filesystem.
func (OSReader) ReadFile(path string) ([]byte, error) { return os.ReadFile(path) }

type (
	// Local selects local typed YAML and optionally SOPS-encrypted secret files.
	Local struct {
		Path        string
		SecretsPath string
		Reader      Reader
	}
	// concealed preserves causes for errors.Is/As while keeping their text out of CLI output.
	concealed struct {
		cause   error
		message string
	}
)

func (e concealed) Error() string { return e.message }
func (e concealed) Unwrap() error { return e.cause }
func hidden(err error, message string) error {
	return fault.Wrap(fault.Configuration, concealed{err, message})
}

// Load reads configuration and encrypted secrets without retaining them in state.
func (s Local) Load(ctx context.Context) (Document, error) {
	var d Document
	if err := ctx.Err(); err != nil {
		return d, err
	}
	reader := s.Reader
	if reader == nil {
		reader = OSReader{}
	}
	if s.Path != "" {
		b, err := reader.ReadFile(s.Path)
		if err != nil {
			return d, hidden(err, "cannot read configuration")
		}
		d.Configuration = b
	}
	if s.SecretsPath != "" {
		b, err := reader.ReadFile(s.SecretsPath)
		if err != nil {
			return Document{}, hidden(err, "cannot read encrypted secrets")
		}
		d.Secrets = b
	}
	return d, nil
}

type (
	// Decrypter isolates secret decryption for deterministic tests.
	Decrypter interface {
		Decrypt(context.Context, []byte) ([]byte, error)
	}
	// SOPS decrypts through the standard host SOPS key environment.
	SOPS struct{}
)

// Decrypt uses SOPS and the host key environment to produce ephemeral plaintext.
func (SOPS) Decrypt(ctx context.Context, b []byte) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	clear, err := decrypt.Data(b, "yaml")
	if err != nil {
		return nil, hidden(err, "SOPS decryption failed")
	}
	if err = ctx.Err(); err != nil {
		clearBytes(clear)
		return nil, err
	}
	return clear, nil
}

func clearBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func decode[T any](b []byte) (T, error) {
	var v T
	if len(bytes.TrimSpace(b)) == 0 {
		return v, nil
	}
	decoder := yaml.NewDecoder(bytes.NewReader(b))
	decoder.KnownFields(true)
	if err := decoder.Decode(&v); err != nil {
		return v, hidden(err, "invalid typed YAML configuration")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return v, hidden(fmt.Errorf("expected one YAML document"), "configuration must contain exactly one YAML document")
	}
	return v, nil
}

// Parse decodes configuration and decrypts secrets into separate user-defined types.
// Call user-specific validation before returning desired resources to the agent.
func Parse[T, S any](ctx context.Context, d Document, decrypter Decrypter) (T, S, error) {
	var zeroT T
	var zeroS S
	v, err := decode[T](d.Configuration)
	if err != nil {
		return zeroT, zeroS, err
	}
	if len(d.Secrets) == 0 {
		return v, zeroS, nil
	}
	if decrypter == nil {
		decrypter = SOPS{}
	}
	clear, err := decrypter.Decrypt(ctx, d.Secrets)
	if err != nil {
		return zeroT, zeroS, hidden(err, "cannot decrypt secrets")
	}
	defer clearBytes(clear)
	secrets, err := decode[S](clear)
	if err != nil {
		return zeroT, zeroS, err
	}
	return v, secrets, nil
}
