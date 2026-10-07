package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const redirectURI = "skydock://callback"

func randomURLSafe(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func s256Challenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func newPKCEPair() (verifier, challenge, state string, err error) {
	verifier, err = randomURLSafe(32)
	if err != nil {
		return "", "", "", fmt.Errorf("generate verifier: %w", err)
	}
	challenge = s256Challenge(verifier)
	state, err = randomURLSafe(32)
	if err != nil {
		return "", "", "", fmt.Errorf("generate state: %w", err)
	}
	return verifier, challenge, state, nil
}
