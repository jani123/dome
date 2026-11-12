package testutil

import (
	"crypto/rand"
	"errors"
	"time"

	"filippo.io/age"
	"github.com/getsops/sops/v3"
	"github.com/getsops/sops/v3/aes"
	sopsage "github.com/getsops/sops/v3/age"
	sopsconfig "github.com/getsops/sops/v3/config"
	"github.com/getsops/sops/v3/stores/yaml"
)

// EncryptYAML generates a fresh host identity and a real SOPS document for tests.
func EncryptYAML(plain []byte) ([]byte, string, error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, "", err
	}
	master, err := sopsage.MasterKeyFromRecipient(identity.Recipient().String())
	if err != nil {
		return nil, "", err
	}
	key := make([]byte, 32)
	if _, err = rand.Read(key); err != nil {
		return nil, "", err
	}
	if err = master.Encrypt(key); err != nil {
		return nil, "", err
	}
	store := yaml.NewStore(&sopsconfig.YAMLStoreConfig{Indent: 2})
	branches, err := store.LoadPlainFile(plain)
	if err != nil {
		return nil, "", err
	}
	tree := sops.Tree{Branches: branches, Metadata: sops.Metadata{LastModified: time.Now().UTC().Truncate(time.Second), Version: "3.13.3", KeyGroups: []sops.KeyGroup{{master}}, UnencryptedSuffix: "_unencrypted"}}
	cipher := aes.NewCipher()
	mac, err := tree.Encrypt(key, cipher)
	if err != nil {
		return nil, "", err
	}
	encryptedMAC, err := cipher.Encrypt(mac, key, tree.Metadata.LastModified.Format(time.RFC3339))
	if err != nil {
		return nil, "", err
	}
	tree.Metadata.MessageAuthenticationCode = encryptedMAC
	b, err := store.EmitEncryptedFile(tree)
	clear(key)
	if err != nil {
		return nil, "", errors.New("cannot emit SOPS fixture")
	}
	return b, identity.String(), nil
}
