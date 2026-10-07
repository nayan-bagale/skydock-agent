package auth

import (
	"github.com/zalando/go-keyring"
)

const keyringService = "SkyDock Agent"
const keyringAccount = "refresh-token"

// KeyStore persists the refresh token in the OS credential store.
type KeyStore interface {
	Set(token string) error
	Get() (string, error)
	Delete() error
}

type keyringStore struct{}

func NewKeyStore() KeyStore {
	return keyringStore{}
}

func (keyringStore) Set(token string) error {
	return keyring.Set(keyringService, keyringAccount, token)
}

func (keyringStore) Get() (string, error) {
	return keyring.Get(keyringService, keyringAccount)
}

func (keyringStore) Delete() error {
	return keyring.Delete(keyringService, keyringAccount)
}
