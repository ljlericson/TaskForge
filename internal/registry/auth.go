package registry

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
)

func (r *Registry) AuthenticateWorker(workerID string, message, signature []byte) error {
	r.pubKeyMutex.RLock()
	pub, ok := r.serverPublicKeys[workerID]
	r.pubKeyMutex.RUnlock()

	if !ok {
		return errors.New("worker public key not found in server registry")
	}

	hash := sha256.Sum256(message)

	err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, hash[:], signature)
	return err
}
