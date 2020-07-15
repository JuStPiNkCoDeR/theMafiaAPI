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
	"crypto/x509"
	"encoding/pem"
	"hash"
)

// RSA data struct
type RSA struct {
	privateKeyOAEP       *rsa.PrivateKey
	privateKeyPSS        *rsa.PrivateKey
	OwnPublicKeyOAEP     *rsa.PublicKey // Public key of server for encryption
	OwnPublicKeyPSS      *rsa.PublicKey // Public key of server for signature
	ForeignPublicKeyOAEP *rsa.PublicKey // Public key of other user to encode data
	ForeignPublicKeyPSS  *rsa.PublicKey // Public key of other user to make signature
	Hash                 hash.Hash      // Hash code
	signHash             crypto.Hash
	signatureOptions     *rsa.PSSOptions
}

// Generate private, public keys and sign hash, options for current instance
func (R *RSA) Init() (err error) {
	if R.privateKeyOAEP, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
		return lib.Wrap(err, "An error occurred while generating keys(OAEP)")
	}

	if R.privateKeyPSS, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
		return lib.Wrap(err, "An error occurred while generating keys(PSS)")
	}

	R.OwnPublicKeyOAEP = &R.privateKeyOAEP.PublicKey
	R.OwnPublicKeyPSS = &R.privateKeyPSS.PublicKey
	R.signatureOptions = &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto}
	R.signHash = crypto.SHA256
	R.Hash = sha256.New()

	return nil
}

// Encode message
func (R *RSA) Encode(message string, label string) (encrypted []byte, err error) {
	if R.ForeignPublicKeyOAEP == nil {
		return nil, lib.Wrap(nil, "Foreign public key(OAEP) is empty")
	}

	encrypted, err = rsa.EncryptOAEP(R.Hash, rand.Reader, R.ForeignPublicKeyOAEP, []byte(message), []byte(label))

	if err != nil {
		return nil, lib.Wrap(err, "An error occurred while encrypting message")
	}

	return
}

// Calculate hash for make/verify sign
func (R *RSA) calculateHashPSS(encrypted []byte) (hash []byte, err error) {
	hashPSS := R.signHash.New()
	_, err = hashPSS.Write(encrypted)

	if err != nil {
		return nil, lib.Wrap(err, "Error on calculating sign PSS hash")
	}

	hash = hashPSS.Sum(nil)

	return
}

// Make a sign for encrypted message
func (R *RSA) Sign(encrypted []byte) (signature []byte, err error) {
	if encrypted == nil {
		return nil, lib.Wrap(nil, "No encrypted string for signature")
	}

	hashed, err := R.calculateHashPSS(encrypted)

	if err != nil {
		return nil, lib.Wrap(err, "Error on making signature")
	}

	signature, err = rsa.SignPSS(rand.Reader, R.privateKeyPSS, R.signHash, hashed, R.signatureOptions)

	if err != nil {
		return nil, lib.Wrap(err, "Error on making signature")
	}

	return
}

// Decode message encoded via algorithm above
func (R *RSA) Decode(encrypted []byte, label []byte) (message string, err error) {
	var bytes []byte
	bytes, err = rsa.DecryptOAEP(R.Hash, rand.Reader, R.privateKeyOAEP, encrypted, label)

	if err != nil {
		return "", lib.Wrap(err, "Error on decoding")
	}

	message = string(bytes)

	return
}

// Check if foreign public key is signs owner key
func (R *RSA) VerifySign(sign []byte, encrypted []byte) (err error) {
	if R.ForeignPublicKeyPSS == nil {
		return lib.Wrap(nil, "Foreign public key(PSS) is empty")
	}

	hashed, err := R.calculateHashPSS(encrypted)

	if err != nil {
		return lib.Wrap(err, "Error on verifying of signature")
	}

	err = rsa.VerifyPSS(R.ForeignPublicKeyPSS, R.signHash, hashed, sign, R.signatureOptions)

	if err != nil {
		return lib.Wrap(err, "Error on verifying signature")
	}

	return nil
}

// Import RSA key from PEM string
func (R *RSA) ImportKey(keyPEM string, isOAEP bool) error {
	block, _ := pem.Decode([]byte(keyPEM))

	if block == nil {
		return &lib.StackError{Message: "Cannot decode PEM string"}
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)

	if err != nil {
		return &lib.StackError{
			ParentError: err,
			Message:     "Cannot parse bytes of rsa key",
		}
	}

	switch pub := key.(type) {
	case *rsa.PublicKey:
		if isOAEP {
			R.ForeignPublicKeyOAEP = pub
		} else {
			R.ForeignPublicKeyPSS = pub
		}

		return nil
	default:
		return &lib.StackError{
			Message: "The input PEM string didnt match rsa public key type",
		}
	}
}

// Export the own public keys as PEM strings
func (R *RSA) ExportKeys() (pemOAEP string, pemPSS string, err error) {
	bytesOAEP, err := x509.MarshalPKIXPublicKey(R.OwnPublicKeyOAEP)

	if err != nil {
		err = lib.Wrap(err, "Error on x509 marshal for OAEP public key")
		return
	}

	pemOAEP = string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: bytesOAEP,
	}))

	if pemOAEP == "" {
		err = lib.Wrap(nil, "Error on encoding public key(OAEP)")
		return
	}

	bytesPSS, err := x509.MarshalPKIXPublicKey(R.OwnPublicKeyPSS)

	if err != nil {
		err = lib.Wrap(err, "Error on x509 marshal for PSS public key")
		return
	}

	pemPSS = string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: bytesPSS,
	}))

	if pemPSS == "" {
		err = lib.Wrap(nil, "Error on encoding public key(PSS)")
		return
	}

	return
}
