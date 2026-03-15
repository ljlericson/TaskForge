// Package registry
package registry

import (
	"crypto/ed25519"
	"fmt"
)

type WorkerConfig struct {
	ID     string `yaml:"id"`
	PubKey string `yaml:"pubkey"`
}

var serverPublicKeys = map[string]ed25519.PublicKey{}

func InitRegistry(wc []WorkerConfig) {
	for _, val := range wc {
		serverPublicKeys[val.ID] = ed25519.PublicKey(val.PubKey)
	}
	fmt.Println(len(serverPublicKeys))
	panic("d")
}

func AuthenticateWorker(workerID string, message, signature []byte) bool {
	pub, ok := serverPublicKeys[workerID]
	if !ok {
		// Unknown worker
		return false
	}

	return ed25519.Verify(pub, message, signature)
}
