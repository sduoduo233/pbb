package update

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/ssh"
)

// Sign a binary using private key.
// Returns signature
func Sign(binaryBytes []byte, privateKeyPem string) ([]byte, error) {

	binaryHash := sha256.Sum256(binaryBytes)

	privateKey, err := ssh.ParseRawPrivateKey([]byte(privateKeyPem))
	if err != nil {
		return nil, fmt.Errorf("bad private key: %w", err)
	}

	rsaPrivkey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("rsa expected")
	}

	signature, err := rsa.SignPKCS1v15(nil, rsaPrivkey, crypto.SHA256, binaryHash[:])
	if err != nil {
		return nil, fmt.Errorf("rsa sign: %w", err)
	}

	return signature, nil
}

// Verify signature.
// Returns not-nil error if the signature does not match
func Verify(binaryBytes []byte, publicKeyPem string, signBytes []byte) error {
	binaryHash := sha256.Sum256(binaryBytes)

	p, _ := pem.Decode([]byte(publicKeyPem))
	if p == nil {
		return fmt.Errorf("bad public key")
	}

	pub, err := x509.ParsePKIXPublicKey(p.Bytes)
	if err != nil {
		return fmt.Errorf("bad public key: %w", err)
	}

	rsaPubkey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("rsa expected")
	}

	return rsa.VerifyPKCS1v15(rsaPubkey, crypto.SHA256, binaryHash[:], signBytes)
}
