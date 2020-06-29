package tests

import (
	"../utils"
	"testing"
)

var Sasha = &utils.RSA{}
var Alisa = &utils.RSA{}

func TestRSA_Init(t *testing.T) {
	err := Sasha.Init()

	if err != nil {
		t.Error("Got error on Sasha RSA init", err)
		t.Fail()
	}

	err = Alisa.Init()

	if err != nil {
		t.Error("Got error on Alisa RSA init", err)
		t.Fail()
	}
}
