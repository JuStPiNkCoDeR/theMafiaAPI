package utils

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"
	"hash"
)

type RSA struct {
	privateKey       *rsa.PrivateKey
	OwnPublicKey     *rsa.PublicKey
	ForeignPublicKey *rsa.PublicKey
	Hash             hash.Hash
	signHash         crypto.Hash
}

// RSA error
type Error struct {
	ParentError error  // Error which has been thrown
	Message     string // String contains short information about error
}

// Return complete error message
func (e *Error) Error() string {
	return fmt.Sprintf("%s occurs for database connection with\n", e.Message)
}

// Generate private and public keys for current instance
func (R *RSA) GenerateOwnKeys() (err error) {
	if R.privateKey, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
		return
	}

	R.OwnPublicKey = &R.privateKey.PublicKey

	return nil
}

func (R *RSA) Encode(message string) (err error, encrypted []byte) {
	if R.ForeignPublicKey == nil {
		return error("213"), []byte("")
	}
}
