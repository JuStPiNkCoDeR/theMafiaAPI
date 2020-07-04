package tests

import (
	"../utils"
	"fmt"
	"testing"
)

var Sasha = &utils.RSA{}
var Alisa = &utils.RSA{}
var message = "Hello, World!"
var secretText []byte
var signature []byte

func TestRSA_Init(t *testing.T) {
	if err := Sasha.Init(); err != nil {
		t.Error("Got error on Sasha RSA init", err)
	}

	if err := Alisa.Init(); err != nil {
		t.Error("Got error on Alisa RSA init", err)
	}

	Sasha.ForeignPublicKey = Alisa.OwnPublicKey
	Alisa.ForeignPublicKey = Sasha.OwnPublicKey
}

func TestRSA_Encode(t *testing.T) {
	if encrypted, err := Sasha.Encode(message, ""); err != nil {
		t.Errorf("Got error on encoding message\n%e", err)
	} else {
		t.Logf("Encrypted: %x", encrypted)
		secretText = encrypted
	}
}

func TestRSA_Decode(t *testing.T) {
	if decoded, err := Alisa.Decode(secretText, []byte("")); err != nil {
		t.Errorf("Got error on decoding message\n%e", err)
	} else {
		if decoded != message {
			t.Errorf("Expected: %s\nGot: %s", message, decoded)
		} else {
			t.Logf("Decoded: %s", decoded)
		}
	}
}

func TestRSA_Sign(t *testing.T) {
	if sign, err := Sasha.Sign(secretText); err != nil {
		t.Errorf("Got error on making signature\n%e", err)
	} else {
		t.Logf("Sign: %x", sign)
		signature = sign
	}
}

func TestRSA_Verify(t *testing.T) {
	if err := Alisa.VerifySign(signature, secretText); err != nil {
		t.Errorf("Got error on verifying signature\n%e", err)
	} else {
		t.Logf("Signature verified")
	}
}

func ExampleRSA_Encode() {
	text := "Hello)"
	var _ []byte

	// Label means short explanation of input
	if encrypted, err := Sasha.Encode(text, ""); err == nil {
		_ = encrypted
	}
}

func ExampleRSA_Decode() {
	text := "Hello)"
	var secret []byte

	// Label means short explanation of input
	if encrypted, err := Sasha.Encode(text, ""); err == nil {
		secret = encrypted
	} else {
		panic("Cant encode!")
	}

	if decoded, err := Alisa.Decode(secret, []byte("")); err == nil {
		fmt.Print(decoded)
	}
	// Output: Hello)
}
