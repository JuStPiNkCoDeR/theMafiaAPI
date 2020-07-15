// Copyright sasha.los.0148@gmail.com
// All Rights have been taken by Mafia :)

// Socket types
package utils

import (
	"../database"
	"../lib"
	"../logger"

	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

var fmtLogger = &logger.MafiaLogger{IsEnabled: true}

type OutEventName string

const (
	rsaSendServerKeys   OutEventName = "rsa:serverKeys"
	rsaAcceptClientKeys OutEventName = "rsa:acceptClientKeys"
	rsaSignUp           OutEventName = "rsa:signUp"
)

// Log functions
func debugSocketEvent(outStream logger.Logger, data interface{}, socket *Socket) {
	dataJSON, err := json.Marshal(data)

	if err != nil {
		outStream.Log(logger.Error, fmt.Sprintf("Unable to serialize data for debug\nError:\n\t%s", err), socket.LogKey)
		return
	}

	outStream.Log(logger.Debug, fmt.Sprintf("Debug socket event\nData:\n\t%s", string(dataJSON)), socket.LogKey)
}

func triggerSocketEvent(outStream logger.Logger, eventName string, socket *Socket) {
	outStream.Log(logger.Debug, fmt.Sprintf("Triggered socket event\n\tNamespace: %s\n\tEvent name: %s\n\tID: %s", socket.Namespace, eventName, socket.ID), socket.LogKey)
}

func successfulSocketEvent(outStream logger.Logger, eventName string, socket *Socket) {
	outStream.Log(logger.Debug, fmt.Sprintf("Socket event passed OK\n\tNamespace: %s\n\tEvent name: %s\n\tID: %s", socket.Namespace, eventName, socket.ID), socket.LogKey)
}

func errorSocketEvent(outStream logger.Logger, eventName string, err error, socket *Socket) {
	outStream.Log(logger.Error, fmt.Sprintf("Error on socket event\n\tNamespace: %s\n\tEvent name: %s\n\tID: %s\n\tError:\n\t%s", socket.Namespace, eventName, socket.ID, err.Error()), socket.LogKey)
}

// Events
func getServerKeys(eventName string, socket *SecureSocket) {
	triggerSocketEvent(fmtLogger, eventName, socket.Socket)

	if err := socket.SendPublicKeys(); err != nil {
		errorSocketEvent(fmtLogger, eventName, err, socket.Socket)
		return
	}

	successfulSocketEvent(fmtLogger, eventName, socket.Socket)
}

func setClientsKeys(incomingMessageMap map[string]interface{}, eventName string, socket *SecureSocket) bool {
	triggerSocketEvent(fmtLogger, eventName, socket.Socket)

	incomingData := incomingMessageMap["data"].(map[string]interface{})

	// Parse OAEP PEM data
	if pemOAEP, ok := incomingData["oaep"]; ok {
		switch typedPEM := pemOAEP.(type) {
		case string:
			if err := socket.r.ImportKey(typedPEM, true); err != nil {
				errorSocketEvent(
					fmtLogger, eventName,
					&lib.StackError{
						ParentError: err,
						Message:     "Error on import OAEP key",
					}, socket.Socket)

				return false
			}

			if socket.r.ForeignPublicKeyOAEP == nil {
				fmtLogger.Log(logger.Warn, "Foreign OAEP rsa key is nil after import", socket.LogKey)
			}
		default:
			errorSocketEvent(fmtLogger, eventName, &lib.StackError{
				Message: "PEM string is not string!",
			}, socket.Socket)

			return false
		}

	}

	// Parse PSS PEM data
	if pemPSS, ok := incomingData["pss"]; ok {
		switch typedPEM := pemPSS.(type) {
		case string:
			if err := socket.r.ImportKey(typedPEM, false); err != nil {
				errorSocketEvent(
					fmtLogger, eventName,
					&lib.StackError{
						ParentError: err,
						Message:     "Error on import PSS key",
					}, socket.Socket)

				return false
			}

			if socket.r.ForeignPublicKeyOAEP == nil {
				fmtLogger.Log(logger.Warn, "Foreign PSS rsa key is nil after import", socket.LogKey)
			}
		default:
			errorSocketEvent(fmtLogger, eventName, &lib.StackError{
				Message: "PEM string is not string!",
			}, socket.Socket)

			return false
		}
	}

	successfulSocketEvent(fmtLogger, eventName, socket.Socket)

	return true
}

func signUp(incomingMessageMap map[string]interface{}, eventName string, socket *SecureSocket) bool {
	triggerSocketEvent(fmtLogger, eventName, socket.Socket)

	// Check if RSA-PSS clients key presents
	if socket.r.ForeignPublicKeyPSS == nil {
		errorSocketEvent(fmtLogger, eventName, &lib.StackError{
			Message: "No foreign public RSA-PSS key present",
		}, socket.Socket)

		return false
	}

	// Check if own private RSA-OAEP key presents
	if socket.r.privateKeyOAEP == nil {
		errorSocketEvent(fmtLogger, eventName, &lib.StackError{
			Message: "No private RSA-OAEP key present",
		}, socket.Socket)

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
			errorSocketEvent(fmtLogger, eventName, &lib.StackError{
				Message: "Can't decode base64 encoded string(Field name: `password`)",
			}, socket.Socket)

			return false
		}

		encodedPassword = buff
	} else {
		errorSocketEvent(fmtLogger, eventName, &lib.StackError{
			Message: "No clients' password presented at incoming data",
		}, socket.Socket)

		return false
	}

	// Get name base64 string and decode it
	if name, ok := incomingData["name"]; ok {
		buff, err := base64.StdEncoding.DecodeString(name.(string))

		if err != nil {
			errorSocketEvent(fmtLogger, eventName, &lib.StackError{
				Message: "Can't decode base64 encoded string(Field name: `name`)",
			}, socket.Socket)

			return false
		}

		encodedUsername = buff
	} else {
		errorSocketEvent(fmtLogger, eventName, &lib.StackError{
			Message: "No clients' name presented at incoming data",
		}, socket.Socket)

		return false
	}

	// Get password signature base64 string and decode it
	if sign, ok := incomingData["passwordSign"]; ok {
		buff, err := base64.StdEncoding.DecodeString(sign.(string))

		if err != nil {
			errorSocketEvent(fmtLogger, eventName, &lib.StackError{
				Message: "Can't decode base64 encoded string(Field name: `signPassword`)",
			}, socket.Socket)

			return false
		}

		signPassword = buff
	} else {
		errorSocketEvent(fmtLogger, eventName, &lib.StackError{
			Message: "No clients' sign for password presented at incoming data",
		}, socket.Socket)

		return false
	}

	// Get name signature base64 string and decode it
	if sign, ok := incomingData["nameSign"]; ok {
		buff, err := base64.StdEncoding.DecodeString(sign.(string))

		if err != nil {
			errorSocketEvent(fmtLogger, eventName, &lib.StackError{
				Message: "Can't decode base64 encoded string(Field name: `signName`)",
			}, socket.Socket)

			return false
		}

		signUsername = buff
	} else {
		errorSocketEvent(fmtLogger, eventName, &lib.StackError{
			Message: "No clients' sign for name presented at incoming data",
		}, socket.Socket)

		return false
	}

	// Verify username signature
	if err := socket.r.VerifySign(signUsername, encodedUsername); err != nil {
		errorSocketEvent(fmtLogger, eventName, &lib.StackError{
			ParentError: err,
			Message:     "Error on verification of name sign",
		}, socket.Socket)

		return false
	}

	// Verify password signature
	if err := socket.r.VerifySign(signPassword, encodedPassword); err != nil {
		errorSocketEvent(fmtLogger, eventName, &lib.StackError{
			ParentError: err,
			Message:     "Error on verification of password sign",
		}, socket.Socket)

		return false
	}

	// Decode username
	if name, err := socket.r.Decode(encodedUsername, []byte("")); err == nil {
		username = name
	} else {
		errorSocketEvent(fmtLogger, eventName, &lib.StackError{
			ParentError: err,
			Message:     "Error on decoding username",
		}, socket.Socket)

		return false
	}

	// Decode password
	if pass, err := socket.r.Decode(encodedPassword, []byte("")); err == nil {
		password = pass
	} else {
		errorSocketEvent(fmtLogger, eventName, &lib.StackError{
			ParentError: err,
			Message:     "Error on decoding password",
		}, socket.Socket)

		return false
	}

	// Make hash string from password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)

	if err != nil {
		errorSocketEvent(fmtLogger, eventName, lib.Wrap(err, "Can't create hash from password"), socket.Socket)

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
	}, socket.Socket)

	// Save to the profiles collection
	err = socket.Database.Insert("profiles", []interface{}{profile}, socket.LogKey)

	if err != nil {
		errorSocketEvent(fmtLogger, eventName, &lib.StackError{
			ParentError: err,
			Message:     "Error on saving profile data",
		}, socket.Socket)

		return false
	}

	successfulSocketEvent(fmtLogger, eventName, socket.Socket)

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
	LogKey    string             // Request ID that send client
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
			// Read new message from client
			_, message, err := socket.Client.ReadMessage()

			if err != nil {
				// Close event handler
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
					errorSocketEvent(fmtLogger, "read", err, socket.Socket)
				} else {
					debugSocketEvent(fmtLogger, "Socket close normally", socket.Socket)
				}

				break
			}

			debugSocketEvent(fmtLogger, map[string]interface{}{
				"Namespace":  socket.Namespace,
				"Event name": "read",
				"ID":         socket.ID,
				"Input":      message,
			}, socket.Socket)

			var incomingMessage interface{}

			// Parse JSON string from message
			err = json.Unmarshal(message, &incomingMessage)

			if err != nil {
				errorSocketEvent(fmtLogger, "unmarshal", err, socket.Socket)
			}

			debugSocketEvent(fmtLogger, map[string]interface{}{
				"Namespace":  socket.Namespace,
				"Event name": "unmarshal",
				"ID":         socket.ID,
				"Input":      incomingMessage,
			}, socket.Socket)

			// Get event name and log key
			incomingMessageMap := incomingMessage.(map[string]interface{})
			eventName := incomingMessageMap["name"].(string)
			socket.LogKey = incomingMessageMap["reqID"].(string)

			switch eventName {
			// Client wants to get servers RSA public keys
			case "rsa:getServerKeys":
				getServerKeys(eventName, socket)
			// Client sends own RSA public keys
			case "rsa:setClientKeys":
				if isAccepted := setClientsKeys(incomingMessageMap, eventName, socket); isAccepted {
					if err := socket.Send(rsaAcceptClientKeys, "YES"); err != nil {
						errorSocketEvent(fmtLogger, eventName, lib.Wrap(err, "Error on sending message"), socket.Socket)
					}
				} else {
					if err := socket.Send(rsaAcceptClientKeys, "NO"); err != nil {
						errorSocketEvent(fmtLogger, eventName, lib.Wrap(err, "Error on sending message"), socket.Socket)
					}
				}
			// Client signing up
			case "rsa:signUp":
				if isSaved := signUp(incomingMessageMap, eventName, socket); isSaved {
					if err := socket.SendEncryptedMessage(rsaSignUp, "YES"); err != nil {
						errorSocketEvent(fmtLogger, eventName, lib.Wrap(err, "Error on sending encrypted message"), socket.Socket)
					}
				} else {
					if err := socket.SendEncryptedMessage(rsaSignUp, "NO"); err != nil {
						errorSocketEvent(fmtLogger, eventName, lib.Wrap(err, "Error on sending encrypted message"), socket.Socket)
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
	pemOAEP, pemPSS, err := socket.r.ExportKeys()

	if err != nil {
		return lib.Wrap(err, "Can't export the own public keys")
	}

	debugSocketEvent(fmtLogger, map[string]interface{}{
		"Event name": "Encode OAEP key to pem",
		"PEM":        pemOAEP,
	}, socket.Socket)

	debugSocketEvent(fmtLogger, map[string]interface{}{
		"Event name": "Encode PSS key to pem",
		"PEM":        pemPSS,
	}, socket.Socket)

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
