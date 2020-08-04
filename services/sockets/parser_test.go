package main

import (
	"encoding/json"
	"testing"
)

func TestSocketParser_Parse(t *testing.T) {
	clientRsa := &RSA{}
	serverRsa := &RSA{}

	if err := clientRsa.Init(); err != nil {
		panic(err)
	}

	if err := serverRsa.Init(); err != nil {
		panic(err)
	}

	serverRsa.ForeignPublicKeyOAEP = clientRsa.OwnPublicKeyOAEP
	serverRsa.ForeignPublicKeyPSS = clientRsa.OwnPublicKeyPSS
	clientRsa.ForeignPublicKeyOAEP = serverRsa.OwnPublicKeyOAEP
	clientRsa.ForeignPublicKeyPSS = serverRsa.OwnPublicKeyPSS

	parser := &SocketParser{rsa: serverRsa}

	type Request struct {
		Name    string `socket:"name,require"`
		Surname string `socket:"surname,omitempty"`
		Email   string `socket:"email,email"`
	}

	type RsaRequest struct {
		Name  string `socket:"name,rsa"`
		Email string `socket:"email,rsa,email"`
	}

	tests := []struct {
		name  string
		rules interface{}
		input DataObject
		want  DataObject
	}{
		{
			name:  "Simple correct request",
			rules: &Request{},
			input: DataObject{
				"name":    "aaa",
				"surname": "bbb",
				"email":   "a@a",
			},
			want: DataObject{
				"name":    "aaa",
				"surname": "bbb",
				"email":   "a@a",
			},
		},
		{
			name:  "Without surname correct request",
			rules: &Request{},
			input: DataObject{
				"name":  "aaa",
				"email": "a@a",
			},
			want: DataObject{
				"name":  "aaa",
				"email": "a@a",
			},
		},
		{
			name:  "Wrong email request",
			rules: &Request{},
			input: DataObject{
				"name":  "aaa",
				"email": "aaa",
			},
			want: nil,
		},
		{
			name:  "Without name wrong request",
			rules: &Request{},
			input: DataObject{
				"email": "a@a",
			},
			want: nil,
		},
		{
			name:  "Without email wrong request",
			rules: &Request{},
			input: DataObject{
				"name": "aaa",
			},
			want: nil,
		},
	}

	for _, test := range tests {
		shouldThrown := test.want == nil
		output := test.rules

		t.Run(test.name, func(t *testing.T) {
			err := parser.Parse(test.input, output)
			if err != nil && !shouldThrown {
				t.Fatalf("Unable to parse data:\n%s", err)
			} else if err == nil && shouldThrown {
				inputJSON, err := json.Marshal(test.input)
				if err != nil {
					panic(err)
				}
				t.Fatalf("Parser didn't catch incorrect data\nInput: %s", inputJSON)
			}

			if !shouldThrown {
				wantJSON, err := json.Marshal(test.want)
				if err != nil {
					panic(err)
				}

				parsedJSON, err := json.Marshal(output)
				if err != nil {
					panic(err)
				}

				if string(wantJSON) == string(parsedJSON) {
					t.Fatalf("Expected:\n\t%s\nGot:\n\t%s", wantJSON, parsedJSON)
				}
			}
		})
	}
}

func ExampleSocketParser_Parse() {

}
