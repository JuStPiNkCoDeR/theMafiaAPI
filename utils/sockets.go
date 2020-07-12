// Copyright sasha.los.0148@gmail.com
// All Rights have been taken by Mafia :)

// Socket types
package utils

import (
	"../database"
	"../lib"
	"../logger"

	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

var fmtLogger = &logger.MafiaLogger{IsEnabled: true}

// Log functions
func debugSocketEvent(outStream logger.Logger, data interface{}) {
	dataJSON, err := json.Marshal(data)

	if err != nil {
		outStream.Log(logger.Error, fmt.Sprintf("Unable to serialize data for debug\nError:\n\t%s\n", err))
		return
	}

	outStream.Log(logger.Debug, fmt.Sprintf("Debug socket event\nData:\n\t%s\n", string(dataJSON)))
}

func triggerSocketEvent(outStream logger.Logger, nsp string, eventName string, ID string) {
	outStream.Log(logger.Debug, fmt.Sprintf("Triggered socket event\n\tNamespace: %s\n\tEvent name: %s\n\tID: %s\n", nsp, eventName, ID))
}

func successfulSocketEvent(outStream logger.Logger, nsp string, eventName string, ID string) {
	outStream.Log(logger.Debug, fmt.Sprintf("Socket event passed OK\n\tNamespace: %s\n\tEvent name: %s\n\tID: %s\n", nsp, eventName, ID))
}

func errorSocketEvent(outStream logger.Logger, nsp string, eventName string, ID string, err error) {
	outStream.Log(logger.Error, fmt.Sprintf("Error on socket event\n\tNamespace: %s\n\tEvent name: %s\n\tID: %s\n\tError:\n\t%s\n", nsp, eventName, ID, err.Error()))
}

type OutEventName string

const (
	rsaSendServerKeys   OutEventName = "rsa:serverKeys"
	rsaAcceptClientKeys OutEventName = "rsa:acceptClientKeys"
	rsaSignUp           OutEventName = "rsa:signUp"
)

// Events
func getServerKeys(nsp string, eventName string, socket *SecureSocket) {
	triggerSocketEvent(fmtLogger, nsp, eventName, socket.ID)

	if err := socket.SendPublicKeys(); err != nil {
		errorSocketEvent(fmtLogger, nsp, eventName, socket.ID, err)
		return
	}

	successfulSocketEvent(fmtLogger, nsp, eventName, socket.ID)
}

func setClientsKeys(incomingMessageMap map[string]interface{}, nsp string, eventName string, socket *SecureSocket) {
	triggerSocketEvent(fmtLogger, nsp, eventName, socket.ID)

	incomingData := incomingMessageMap["data"].(map[string]interface{})

	// Parse OAEP PEM data
	if pemOAEP, ok := incomingData["oaep"]; ok {
		switch typedPEM := pemOAEP.(type) {
		case string:
			if err := socket.r.ImportKey(typedPEM, true); err != nil {
				errorSocketEvent(
					fmtLogger,
					nsp,
					eventName,
					socket.ID,
					&lib.StackError{
						ParentError: err,
						Message:     "Error on import OAEP key",
					})

				if err := socket.Send(rsaAcceptClientKeys, "NO"); err != nil {
					errorSocketEvent(fmtLogger, nsp, eventName, socket.ID, err)
				}

				return
			}

			if socket.r.ForeignPublicKeyOAEP == nil {
				fmtLogger.Log(logger.Warn, "Foreign OAEP rsa key is nil after import")
			}
		default:
			errorSocketEvent(fmtLogger, nsp, eventName, socket.ID, &lib.StackError{
				Message: "PEM string is not string!",
			})

			if err := socket.Send(rsaAcceptClientKeys, "NO"); err != nil {
				errorSocketEvent(fmtLogger, nsp, eventName, socket.ID, err)
			}

			return
		}

	}

	// Parse PSS PEM data
	if pemPSS, ok := incomingData["pss"]; ok {
		switch typedPEM := pemPSS.(type) {
		case string:
			if err := socket.r.ImportKey(typedPEM, false); err != nil {
				errorSocketEvent(
					fmtLogger,
					nsp,
					eventName,
					socket.ID,
					&lib.StackError{
						ParentError: err,
						Message:     "Error on import PSS key",
					})

				if err := socket.Send(rsaAcceptClientKeys, "NO"); err != nil {
					errorSocketEvent(fmtLogger, nsp, eventName, socket.ID, err)
				}

				return
			}

			if socket.r.ForeignPublicKeyOAEP == nil {
				fmtLogger.Log(logger.Warn, "Foreign PSS rsa key is nil after import")
			}
		default:
			errorSocketEvent(fmtLogger, nsp, eventName, socket.ID, &lib.StackError{
				Message: "PEM string is not string!",
			})

			if err := socket.Send(rsaAcceptClientKeys, "NO"); err != nil {
				errorSocketEvent(fmtLogger, nsp, eventName, socket.ID, err)
			}

			return
		}
	}

	if err := socket.Send(rsaAcceptClientKeys, "YES"); err != nil {
		errorSocketEvent(fmtLogger, socket.Namespace, eventName, socket.ID, err)
		return
	}

	successfulSocketEvent(fmtLogger, socket.Namespace, eventName, socket.ID)
}

func signUp(incomingMessageMap map[string]interface{}, nsp string, eventName string, socket *SecureSocket) bool {
	triggerSocketEvent(fmtLogger, nsp, eventName, socket.ID)

	// Check if RSA-PSS clients key presents
	if socket.r.ForeignPublicKeyPSS == nil {
		errorSocketEvent(fmtLogger, nsp, eventName, socket.ID, &lib.StackError{
			Message: "No foreign public RSA-PSS key present",
		})

		return false
	}

	// Check if own private RSA-OAEP key presents
	if socket.r.privateKeyOAEP == nil {
		errorSocketEvent(fmtLogger, nsp, eventName, socket.ID, &lib.StackError{
			Message: "No private RSA-OAEP key present",
		})

		return false
	}

	var (
		incomingData    = incomingMessageMap["data"].(map[string]interface{})
		encodedUsername []byte
		encodedPassword []byte
		username        string
		password        string
		signUsername    []byte
		signPassword    []byte
	)

	// Get password base64 string and decode it
	if pass, ok := incomingData["password"]; ok {
		buff, err := base64.StdEncoding.DecodeString(pass.(string))

		if err != nil {
			errorSocketEvent(fmtLogger, nsp, eventName, socket.ID, &lib.StackError{
				Message: "Can't decode base64 encoded string(Field name: `password`)",
			})

			return false
		}

		encodedPassword = buff
	} else {
		errorSocketEvent(fmtLogger, nsp, eventName, socket.ID, &lib.StackError{
			Message: "No clients' password presented at incoming data",
		})

		return false
	}

	// Get name base64 string and decode it
	if name, ok := incomingData["name"]; ok {
		buff, err := base64.StdEncoding.DecodeString(name.(string))

		if err != nil {
			errorSocketEvent(fmtLogger, nsp, eventName, socket.ID, &lib.StackError{
				Message: "Can't decode base64 encoded string(Field name: `name`)",
			})

			return false
		}

		encodedUsername = buff
	} else {
		errorSocketEvent(fmtLogger, nsp, eventName, socket.ID, &lib.StackError{
			Message: "No clients' name presented at incoming data",
		})

		return false
	}

	// Get password signature base64 string and decode it
	if sign, ok := incomingData["signPassword"]; ok {
		buff, err := base64.StdEncoding.DecodeString(sign.(string))

		if err != nil {
			errorSocketEvent(fmtLogger, nsp, eventName, socket.ID, &lib.StackError{
				Message: "Can't decode base64 encoded string(Field name: `signPassword`)",
			})

			return false
		}

		signPassword = buff
	} else {
		errorSocketEvent(fmtLogger, nsp, eventName, socket.ID, &lib.StackError{
			Message: "No clients' sign for password presented at incoming data",
		})

		return false
	}

	// Get name signature base64 string and decode it
	if sign, ok := incomingData["signName"]; ok {
		buff, err := base64.StdEncoding.DecodeString(sign.(string))

		if err != nil {
			errorSocketEvent(fmtLogger, nsp, eventName, socket.ID, &lib.StackError{
				Message: "Can't decode base64 encoded string(Field name: `signName`)",
			})

			return false
		}

		signUsername = buff
	} else {
		errorSocketEvent(fmtLogger, nsp, eventName, socket.ID, &lib.StackError{
			Message: "No clients' sign for name presented at incoming data",
		})

		return false
	}

	// Verify username signature
	if err := socket.r.VerifySign(signUsername, encodedUsername); err != nil {
		errorSocketEvent(fmtLogger, nsp, eventName, socket.ID, &lib.StackError{
			ParentError: err,
			Message:     "Error on verification of name sign",
		})

		return false
	}

	// Verify password signature
	if err := socket.r.VerifySign(signPassword, encodedPassword); err != nil {
		errorSocketEvent(fmtLogger, nsp, eventName, socket.ID, &lib.StackError{
			ParentError: err,
			Message:     "Error on verification of password sign",
		})

		return false
	}

	// Decode username
	if name, err := socket.r.Decode(encodedUsername, []byte("")); err == nil {
		username = name
	} else {
		errorSocketEvent(fmtLogger, nsp, eventName, socket.ID, &lib.StackError{
			ParentError: err,
			Message:     "Error on decoding username",
		})

		return false
	}

	// Decode password
	if pass, err := socket.r.Decode(encodedPassword, []byte("")); err == nil {
		password = pass
	} else {
		errorSocketEvent(fmtLogger, nsp, eventName, socket.ID, &lib.StackError{
			ParentError: err,
			Message:     "Error on decoding password",
		})

		return false
	}

	// Make hash string from password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)

	if err != nil {
		errorSocketEvent(fmtLogger, nsp, eventName, socket.ID, lib.Wrap(err, "Can't create hash from password"))

		return false
	}

	// Create profile object
	profile := database.Profile{
		Name:     username,
		Password: string(hash),
	}

	debugSocketEvent(fmtLogger, map[string]string{
		"Event name": "rsa:signUp",
		"Username":   username,
		"Password":   password,
	})

	// Save to the profiles collection
	err = socket.Database.Insert("profiles", []interface{}{profile})

	if err != nil {
		errorSocketEvent(fmtLogger, nsp, eventName, socket.ID, &lib.StackError{
			ParentError: err,
			Message:     "Error on saving profile data",
		})

		return false
	}

	successfulSocketEvent(fmtLogger, nsp, eventName, socket.ID)

	return true
}

// Socket response struct
type Response struct {
	Name OutEventName `json:"name"`
	Data string       `json:"data"`
}

// Common socket struct
type Socket struct {
	Database  *database.Database // Database connection instance
	Client    *websocket.Conn    // Connection instance
	ID        string             // Socket ID
	Namespace string             // Current socket namespace
}

func (socket *Socket) Send(eventName OutEventName, data string) error {
	response := &Response{
		Name: eventName,
		Data: data,
	}

	resJSON, err := json.Marshal(response)

	if err != nil {
		return &lib.StackError{
			ParentError: err,
			Message:     "Error on marshal response object to json",
		}
	}

	err = socket.Client.WriteMessage(websocket.TextMessage, resJSON)

	if err != nil {
		return &lib.StackError{
			ParentError: err,
			Message:     "Error on writing to socket",
		}
	}

	return nil
}

// Socket that use RSA encryption
type SecureSocket struct {
	r *RSA // RSA object to encode/decode
	*Socket
}

// Init socket
func (socket *SecureSocket) Init() error {
	socket.r = &RSA{}
	socket.ID = socket.Client.RemoteAddr().String()
	socket.Namespace = "secure"

	err := socket.r.Init()

	if err != nil {
		return &lib.StackError{
			ParentError: err,
			Message:     "Error on initializing of secure socket",
		}
	}

	done := make(chan struct{})

	go func() {
		defer close(done)

		for {
			_, message, err := socket.Client.ReadMessage()

			if err != nil {
				errorSocketEvent(fmtLogger, socket.Namespace, "read", socket.ID, err)
				return
			}

			debugSocketEvent(fmtLogger, map[string]interface{}{
				"Namespace":  socket.Namespace,
				"Event name": "read",
				"ID":         socket.ID,
				"Input":      message,
			})

			var incomingMessage interface{}

			err = json.Unmarshal(message, &incomingMessage)

			if err != nil {
				errorSocketEvent(fmtLogger, socket.Namespace, "unmarshal", socket.ID, err)
			}

			debugSocketEvent(fmtLogger, map[string]interface{}{
				"Namespace":  socket.Namespace,
				"Event name": "unmarshal",
				"ID":         socket.ID,
				"Input":      incomingMessage,
			})

			incomingMessageMap := incomingMessage.(map[string]interface{})
			eventName := incomingMessageMap["name"].(string)

			switch eventName {
			case "rsa:getServerKeys":
				getServerKeys(socket.Namespace, eventName, socket)
			case "rsa:setClientKeys":
				setClientsKeys(incomingMessageMap, socket.Namespace, eventName, socket)
			case "rsa:signUp":
				if isSaved := signUp(incomingMessageMap, socket.Namespace, eventName, socket); isSaved {
					if err := socket.SendEncryptedMessage(rsaSignUp, "YES"); err != nil {
						errorSocketEvent(fmtLogger, socket.Namespace, eventName, socket.ID, &lib.StackError{
							ParentError: err,
							Message:     "Error on sending encrypted message",
						})
					}
				} else {
					if err := socket.SendEncryptedMessage(rsaSignUp, "NO"); err != nil {
						errorSocketEvent(fmtLogger, socket.Namespace, eventName, socket.ID, &lib.StackError{
							ParentError: err,
							Message:     "Error on sending encrypted message",
						})
					}
				}
			}
		}
	}()

	return nil
}

// Send encrypted single message
func (socket *SecureSocket) SendEncryptedMessage(eventName OutEventName, data string) error {
	encrypted, err := socket.r.Encode(data, "")

	if err == nil {
		if err := socket.Send(eventName, base64.StdEncoding.EncodeToString(encrypted)); err != nil {
			return err
		}

		return nil
	}

	return &lib.StackError{
		ParentError: err,
		Message:     "Error on socket encryption",
	}
}

// Send RSA public key as JSON string
func (socket *SecureSocket) SendPublicKeys() error {
	bytesOAEP, err := x509.MarshalPKIXPublicKey(socket.r.OwnPublicKeyOAEP)

	if err != nil {
		return &lib.StackError{
			ParentError: err,
			Message:     "Error on x509 marshal for OAEP public key",
		}
	}

	pemOAEP := string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: bytesOAEP,
	}))

	if pemOAEP == "" {
		return &lib.StackError{
			Message: "Error on encoding public key(OAEP)",
		}
	}

	debugSocketEvent(fmtLogger, map[string]interface{}{
		"Event name": "Encode OAEP key to pem",
		"PEM":        pemOAEP,
	})

	bytesPSS, err := x509.MarshalPKIXPublicKey(socket.r.OwnPublicKeyPSS)

	if err != nil {
		return &lib.StackError{
			ParentError: err,
			Message:     "Error on ASN marshal for PSS public key",
		}
	}

	pemPSS := string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: bytesPSS,
	}))

	if pemPSS == "" {
		return &lib.StackError{
			Message: "Error on encoding public key(PSS)",
		}
	}

	debugSocketEvent(fmtLogger, map[string]interface{}{
		"Event name": "Encode PSS key to pem",
		"PEM":        pemPSS,
	})

	data := map[string]interface{}{
		"OAEP": pemOAEP,
		"PSS":  pemPSS,
	}

	dataJSON, err := json.Marshal(data)

	if err != nil {
		return &lib.StackError{
			ParentError: err,
			Message:     "Error on marshal response to json(public keys)",
		}
	}

	err = socket.Send(rsaSendServerKeys, string(dataJSON))

	if err != nil {
		return &lib.StackError{
			ParentError: err,
			Message:     "Error on writing to socket",
		}
	}

	return nil
}
