// Package registry
package registry

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
)

func (r *Registry) AuthenticateWorker(workerID string, message, signature []byte) bool {
	r.pubKeyMutex.RLock()
	pub, ok := r.serverPublicKeys[workerID]
	r.pubKeyMutex.RUnlock()

	if !ok {
		return false
	}

	hash := sha256.Sum256(message)

	err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, hash[:], signature)
	return err == nil
}
