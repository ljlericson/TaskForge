package registry

import (
	"crypto/ed25519"
	"sync"
	"testing"
)

func setupTestRegistry() (string, ed25519.PublicKey, ed25519.PrivateKey) {
	pub, priv, _ := ed25519.GenerateKey(nil)

	registryInstance.pubKeyMutex.Lock()
	registryInstance.serverPublicKeys = map[string]ed25519.PublicKey{}
	registryInstance.serverPublicKeys["worker1"] = pub
	registryInstance.pubKeyMutex.Unlock()

	return "worker1", pub, priv
}

func TestAuthenticateWorker_ValidSignature(t *testing.T) {
	id, _, priv := setupTestRegistry()

	msg := []byte("hello worker")
	sig := ed25519.Sign(priv, msg)

	if !AuthenticateWorker(id, msg, sig) {
		t.Fatal("expected authentication to succeed")
	}
}

func TestAuthenticateWorker_InvalidSignature(t *testing.T) {
	id, _, priv := setupTestRegistry()

	msg := []byte("hello worker")
	sig := ed25519.Sign(priv, msg)

	sig[0] ^= 0xFF

	if AuthenticateWorker(id, msg, sig) {
		t.Fatal("expected authentication to fail")
	}
}

func TestAuthenticateWorker_UnknownWorker(t *testing.T) {
	setupTestRegistry()

	msg := []byte("hello worker")
	sig := make([]byte, 64)

	if AuthenticateWorker("unknownWorker", msg, sig) {
		t.Fatal("expected authentication to fail for unknown worker")
	}
}

func TestAuthenticateWorker_ModifiedMessage(t *testing.T) {
	id, _, priv := setupTestRegistry()

	msg := []byte("original")
	sig := ed25519.Sign(priv, msg)

	modified := []byte("tampered")

	if AuthenticateWorker(id, modified, sig) {
		t.Fatal("expected authentication to fail for modified message")
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

func TestAuthenticateWorker_ConcurrentAccess(t *testing.T) {
	id, _, priv := setupTestRegistry()

	msg := []byte("parallel test")
	sig := ed25519.Sign(priv, msg)

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if !AuthenticateWorker(id, msg, sig) {
				t.Error("expected authentication to succeed concurrently")
			}
		}()
	}

	wg.Wait()
}
