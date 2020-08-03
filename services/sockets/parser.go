// Copyright sasha.los.0148@gmail.com
// All Rights have been taken by Mafia :)
package main

import (
	"mafia/sockets/lib"
	"regexp"
	"strings"

	"encoding/base64"
	"reflect"
)

var (
	emailRegExp = regexp.MustCompile(`^.+@.+$`)
)

type Option string

const (
	tagKey = "socket"
	// Enable decode and verify sign algorithms
	rsa Option = "rsa"
	// Field can be empty(nil)
	omitempty Option = "omitempty"
	// Filed should present at input
	require Option = "require"
	// Field should pass email RegExp
	email Option = "email"
)

// Simple JSON parsed representation
type DataObject map[string]interface{}

/*
Parse & validate input data using giving rules struct.
Use `socket:"..."` tag to set parse options
*/
type Parser interface {
	Parse(input DataObject, rules interface{}) (DataObject, error)
}

// Check the correctness of input type
func checkType(input interface{}, t reflect.Type) bool {
	inType := reflect.Indirect(reflect.ValueOf(input)).Type()

	return inType.AssignableTo(t)
}

// Simple socket parser
type SocketParser struct {
	rsa *RSA
}

// Transform & validate input using options
func (p *SocketParser) lookFor(input DataObject, field reflect.StructField, found bool, option Option) (bool, interface{}, error) {
	data := input[field.Name]

	switch option {
	case omitempty:
		return true, nil, nil
	case require:
		return found && data != nil && checkType(data, field.Type), nil, nil
	case email:
		return found && emailRegExp.MatchString(data.(string)), nil, nil
	case rsa:
		var (
			encrypted   string
			sign        string
			buffSign    []byte
			buffEncoded []byte
			decoded     string
		)

		if p.rsa == nil {
			return false, nil, lib.Wrap(nil, "The RSA object is nil, but required for decryption")
		}

		if !found {
			return false, nil, lib.Wrap(nil, "Unable to find encrypted data")
		}

		encrypted = data.(string)
		sign, ok := input[field.Name+"Sign"].(string)
		if !ok {
			return false, nil, lib.Wrap(nil, "Unable to find RSA sign")
		}

		buffSign, err := base64.StdEncoding.DecodeString(sign)
		if err != nil {
			return false, nil, lib.Wrap(err, "Unable to decode sign string(base64)")
		}

		buffEncoded, err = base64.StdEncoding.DecodeString(encrypted)
		if err != nil {
			return false, nil, lib.Wrap(err, "Unable to decode encoded string(base64)")
		}

		if err := p.rsa.VerifySign(buffSign, buffEncoded); err != nil {
			return false, nil, lib.Wrap(err, "Unable to verify sign")
		}

		decoded, err = p.rsa.Decode(buffEncoded, []byte(""))
		if err != nil {
			return false, nil, lib.Wrap(err, "Unable to decode input")
		}

		return true, decoded, nil
	default:
		return false, nil, lib.Wrap(nil, "Unknown option")
	}
}

// Runs parse logic
func (p *SocketParser) Parse(input DataObject, rules interface{}) (DataObject, error) {
	var (
		val          = reflect.Indirect(reflect.ValueOf(rules))
		fieldsAmount = val.Type().NumField()
		output       = make(DataObject)
	)

	for i := 0; i < fieldsAmount; i++ {
		field := val.Type().Field(i)
		data, ok := input[field.Name]
		// Tag checking
		if field.Tag == "" {
			// No tags -> smooth checking
			if !ok {
				// No rule & no input -> skip
				continue
			}
			// It appears at input -> check type
			if checkType(data, field.Type) {
				// Types assignable -> accept
				output[field.Name] = data
			}
		} else {
			// Tags specified -> hard checking
			optionString := field.Tag.Get(tagKey)
			options := strings.Split(optionString, ",")

			for _, opt := range options {
				isCorrect, parsed, err := p.lookFor(input, field, ok, Option(opt))
				if err != nil {
					return nil, lib.Wrap(err, "Error on parsing input object")
				}

				if !isCorrect {
					return nil, lib.Wrap(nil, "Incorrect data passed for "+opt+" option")
				}

				if parsed != nil {
					input[field.Name] = parsed
				}
			}

			if ok {
				output[field.Name] = input[field.Name]
			}
		}
	}

	return output, nil
}
