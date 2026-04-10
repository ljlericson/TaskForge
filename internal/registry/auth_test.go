package registry

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

func setupTestRegistry() (string, *rsa.PublicKey, *rsa.PrivateKey) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	pub := &priv.PublicKey

	registryInstance.pubKeyMutex.Lock()
	registryInstance.serverPublicKeys = map[string]*rsa.PublicKey{}
	registryInstance.serverPublicKeys["worker1"] = pub
	registryInstance.pubKeyMutex.Unlock()

	return "worker1", pub, priv
}

func TestAuthenticateWorker_UnknownWorker(t *testing.T) {
	setupTestRegistry()

	msg := []byte("hello worker")
	sig := make([]byte, 64)

	if AuthenticateWorker("unknownWorker", msg, sig) {
		t.Fatal("expected authentication to fail for unknown worker")
	}
}

func TestAuthenticateWorker_WrongKey(t *testing.T) {
	id, _, _ := setupTestRegistry()

	_, wrongPriv, _ := ed25519.GenerateKey(nil)

	msg := []byte("hello")
	sig := ed25519.Sign(wrongPriv, msg)

	if AuthenticateWorker(id, msg, sig) {
		t.Fatal("expected authentication to fail with wrong key")
	}
}

func TestAuthenticateWorker_EmptyInputs(t *testing.T) {
	id, _, _ := setupTestRegistry()

	if AuthenticateWorker(id, []byte{}, []byte{}) {
		t.Fatal("expected authentication to fail with empty inputs")
	}
}
