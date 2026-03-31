// Package registry
package registry

import "crypto/ed25519"

func AuthenticateWorker(workerID string, message, signature []byte) bool {
	registryInstance.pubKeyMutex.RLock()
	pub, ok := registryInstance.serverPublicKeys[workerID]
	registryInstance.pubKeyMutex.RUnlock()

	if !ok {
		return false
	}

	return ed25519.Verify(pub, message, signature)
}
