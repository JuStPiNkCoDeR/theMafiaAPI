package tests

import (
	"../utils"
	"testing"
)

var Sasha = &utils.RSA{}
var Alisa = &utils.RSA{}
var message = "Hello, World!"
var secretText []byte

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
