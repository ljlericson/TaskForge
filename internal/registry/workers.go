// Package registry
package registry

import (
	"crypto/ed25519"
	"sync"

	"github.com/ljlericson/TaskForge/internal/console"
)

type WorkerConfig struct {
	ID     string `yaml:"id"`
	PubKey string `yaml:"pubkey"`
}

var (
	pubKeyMutex      = sync.RWMutex{}
	serverPublicKeys = map[string]ed25519.PublicKey{}
)

func InitRegistry(wc []WorkerConfig) {
	pubKeyMutex.Lock()
	for _, val := range wc {
		serverPublicKeys[val.ID] = ed25519.PublicKey(val.PubKey)
	}
	pubKeyMutex.Unlock()
}

func ListWorkers(c *console.Console) {
	pubKeyMutex.RLock()
	for key := range serverPublicKeys {
		c.Log(key)
	}
	pubKeyMutex.RUnlock()
}

func AuthenticateWorker(workerID string, message, signature []byte) bool {
	pubKeyMutex.RLock()
	pub, ok := serverPublicKeys[workerID]
	pubKeyMutex.RUnlock()

	if !ok {
		return false
	}

	return ed25519.Verify(pub, message, signature)
}
