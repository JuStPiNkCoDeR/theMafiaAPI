package main

import (
	"encoding/json"
	"testing"
)

func compare(a, b DataObject) bool {
	for key, value := range a {
		bValue, ok := b[key]
		if !(ok && bValue == value) {
			return false
		}
	}

	return true
}

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
		name    string `socket:"require"`
		surname string `socket:"omitempty"`
		email   string `socket:"email"`
	}

	type RsaRequest struct {
		name  string `socket:"rsa"`
		email string `socket:"rsa,email"`
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

		t.Run(test.name, func(t *testing.T) {
			parsed, err := parser.Parse(test.input, test.rules)
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

				parsedJSON, err := json.Marshal(parsed)
				if err != nil {
					panic(err)
				}

				if !(compare(test.want, parsed) && string(wantJSON) == string(parsedJSON)) {
					t.Fatalf("Expected:\n\t%s\nGot:\n\t%s", wantJSON, parsedJSON)
				}
			}
		})
	}
}

func ExampleSocketParser_Parse() {

}
