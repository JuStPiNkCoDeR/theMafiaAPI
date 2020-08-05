// Copyright sasha.los.0148@gmail.com
// All Rights have been taken by Mafia :)
package main

import (
	"fmt"
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
	Parse(input DataObject, output interface{}) error
}

// Check the correctness of input type
func checkType(input interface{}, t reflect.Type) bool {
	inType := reflect.Indirect(reflect.ValueOf(input)).Type()

	return inType.AssignableTo(t)
}

// Set value directly to named field
func setField(output interface{}, name string, value interface{}) error {
	// Skip nil values
	if value == nil {
		return nil
	}

	structValue := reflect.ValueOf(output).Elem()
	fieldValue := structValue.FieldByName(name)

	if !fieldValue.IsValid() {
		return lib.Wrap(nil, fmt.Sprintf("No such field: %s in struct", name))
	}

	if !fieldValue.CanSet() {
		return lib.Wrap(nil, fmt.Sprintf("Can't set %s field", name))
	}

	fieldType := fieldValue.Type()
	val := reflect.ValueOf(value)

	if fieldType != val.Type() {
		return lib.Wrap(nil, "Provided value didn't match with struct field type")
	}

	fieldValue.Set(val)
	return nil
}

// Simple socket parser
type SocketParser struct {
	rsa *RSA
}

// Transform & validate input using options
func (p *SocketParser) lookFor(input DataObject, inputFieldName string, field reflect.StructField, option Option) (bool, interface{}, error) {
	data, found := input[inputFieldName]

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
func (p *SocketParser) Parse(input DataObject, output interface{}) error {
	var (
		val          = reflect.Indirect(reflect.ValueOf(output))
		fieldsAmount = val.Type().NumField()
	)

	for i := 0; i < fieldsAmount; i++ {
		field := val.Type().Field(i)
		// Tag checking
		if field.Tag == "" {
			data, ok := input[field.Name]
			// No tags -> smooth checking
			if !ok {
				// No rule & no input -> skip
				continue
			}
			// It appears at input -> check type
			if checkType(data, field.Type) {
				// Types assignable -> accept
				if err := setField(output, field.Name, data); err != nil {
					return lib.Wrap(err, "Error occurred while setting field value")
				}
			}
		} else {
			// Tags specified -> hard checking
			var (
				temp           interface{}
				optionString   = field.Tag.Get(tagKey)
				options        = strings.Split(optionString, ",")
				// First option should be input object field name
				inputFieldName = options[0]
			)

			for _, opt := range options[1:] {
				isCorrect, parsed, err := p.lookFor(input, inputFieldName, field, Option(opt))
				if err != nil {
					return lib.Wrap(err, "Error on parsing input object")
				}

				if !isCorrect {
					return lib.Wrap(nil, "Incorrect data passed for "+opt+" option")
				}

				if parsed != nil {
					temp = parsed
				} else {
					temp = input[inputFieldName]
				}
			}

			if err := setField(output, field.Name, temp); err != nil {
				return lib.Wrap(err, "Error occurred while setting field value")
			}
		}
	}

	return nil
}
