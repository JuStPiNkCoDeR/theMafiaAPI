// Copyright sasha.los.0148@gmail.com
// All Rights have been taken by Mafia :)

// RSA encryption implementation
package utils

import (
	"../lib"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"hash"
)

// RSA data struct
type RSA struct {
	privateKey       *rsa.PrivateKey
	OwnPublicKey     *rsa.PublicKey // Public key for current user
	ForeignPublicKey *rsa.PublicKey // Public key for other user
	Hash             hash.Hash      // Hash code
	signHash         crypto.Hash    // Hash code for sign
}

// Generate private and public keys for current instance
func (R *RSA) GenerateOwnKeys() (err error) {
	if R.privateKey, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
		return &lib.StackError{
			ParentError: err,
			Message:     "An error occurred while generating keys",
		}
	}

	R.OwnPublicKey = &R.privateKey.PublicKey

	return nil
}

// Encode message
func (R *RSA) Encode(message string, label string) (encrypted []byte, err error) {
	if R.ForeignPublicKey == nil {
		return nil, &lib.StackError{Message: "Foreign public key is empty"}
	}

	encrypted, err = rsa.EncryptOAEP(R.Hash, rand.Reader, R.ForeignPublicKey, []byte(message), []byte(label))

	if err != nil {
		return nil, &lib.StackError{
			ParentError: err,
			Message:     "An error occurred while encrypting message",
		}
	} else {
		return
	}
}
