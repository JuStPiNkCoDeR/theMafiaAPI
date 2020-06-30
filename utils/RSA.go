// Copyright sasha.los.0148@gmail.com
// All Rights have been taken by Mafia :)

// RSA encryption implementation
package utils

import (
	"../lib"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"hash"
)

// RSA data struct
type RSA struct {
	privateKey       *rsa.PrivateKey
	OwnPublicKey     *rsa.PublicKey // Public key of current user
	ForeignPublicKey *rsa.PublicKey // Public key of other user
	Hash             hash.Hash      // Hash code
	signHash         crypto.Hash
	signatureOptions *rsa.PSSOptions
}

// Generate private, public keys and sign hash, options for current instance
func (R *RSA) Init() (err error) {
	if R.privateKey, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
		return &lib.StackError{
			ParentError: err,
			Message:     "An error occurred while generating keys",
		}
	}

	R.OwnPublicKey = &R.privateKey.PublicKey
	R.signatureOptions = &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto}
	R.signHash = crypto.SHA256
	R.Hash = sha256.New()

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
	}
	return
}

// Calculate hash for make/verify sign
func (R *RSA) calculateHashPSS(encrypted []byte) (hash []byte, err error) {
	hashPSS := R.signHash.New()
	_, err = hashPSS.Write(encrypted)

	if err != nil {
		return nil, &lib.StackError{
			ParentError: err,
			Message:     "Error on calculating sign PSS hash",
		}
	}

	hash = hashPSS.Sum(nil)

	return
}

// Make a sign for encrypted message
func (R *RSA) Sign(encrypted []byte) (signature []byte, err error) {
	if encrypted == nil {
		return nil, &lib.StackError{Message: "No encrypted string for signature"}
	}

	hashed, err := R.calculateHashPSS(encrypted)

	if err != nil {
		return nil, &lib.StackError{
			ParentError: err,
			Message:     "Error on making signature",
		}
	}

	signature, err = rsa.SignPSS(rand.Reader, R.privateKey, R.signHash, hashed, R.signatureOptions)

	if err != nil {
		return nil, &lib.StackError{ParentError: err, Message: "Error on making signature"}
	}

	return
}

// Decode message encoded via algorithm above
func (R *RSA) Decode(encrypted []byte, label []byte) (message string, err error) {
	var bytes []byte
	bytes, err = rsa.DecryptOAEP(R.Hash, rand.Reader, R.privateKey, encrypted, label)

	if err != nil {
		return "", &lib.StackError{ParentError: err, Message: "Error on decoding"}
	}

	message = string(bytes)

	return
}

// Check if foreign public key is signs owner key
func (R *RSA) VerifySign(sign []byte, encrypted []byte) (err error) {
	hashed, err := R.calculateHashPSS(encrypted)

	if err != nil {
		return &lib.StackError{
			ParentError: err,
			Message:     "Error on verifying of signature",
		}
	}

	err = rsa.VerifyPSS(R.ForeignPublicKey, R.signHash, hashed, sign, R.signatureOptions)

	if err != nil {
		return &lib.StackError{ParentError: err, Message: "Error on verifying signature"}
	}

	return nil
}
