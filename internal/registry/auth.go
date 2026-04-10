// Package registry
package registry

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
)

func AuthenticateWorker(workerID string, message, signature []byte) bool {
	registryInstance.pubKeyMutex.RLock()
	pub, ok := registryInstance.serverPublicKeys[workerID]
	registryInstance.pubKeyMutex.RUnlock()

	if !ok {
		return false
	}

	hash := sha256.Sum256(message)

	err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, hash[:], signature)
	return err == nil
}
