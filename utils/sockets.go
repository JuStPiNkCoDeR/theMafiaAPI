// Copyright sasha.los.0148@gmail.com
// All Rights have been taken by Mafia :)

// Socket types
package utils

import (
	"../lib"
	"../logger"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	pem "encoding/pem"
	"fmt"
	"github.com/gorilla/websocket"
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
	outStream.Log(logger.Error, fmt.Sprintf("Error on socket event\n\tNamespace: %s\n\tEvent name: %s\n\tID: %s\n\tError:\n\t\t%s\n", nsp, eventName, ID, err.Error()))
}

type OutEventName string

const (
	rsaSendServerKeys   OutEventName = "rsa:serverKeys"
	rsaAcceptClientKeys OutEventName = "rsa:acceptClientKeys"
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
		errorSocketEvent(fmtLogger, "secure", eventName, socket.ID, err)
		return
	}

	successfulSocketEvent(fmtLogger, "secure", eventName, socket.ID)
}

// Socket response struct
type Response struct {
	Name OutEventName `json:"name"`
	Data string       `json:"data"`
}

// Socket that use RSA encryption
type SecureSocket struct {
	r      *RSA            // RSA object to encode/decode
	Client *websocket.Conn // Connection instance
	ID     string          // Socket ID
}

// Init socket
func (socket *SecureSocket) Init() error {
	socket.r = &RSA{}
	socket.ID = socket.Client.RemoteAddr().String()

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
				errorSocketEvent(fmtLogger, "secure", "read", socket.ID, err)
				return
			}

			debugSocketEvent(fmtLogger, map[string]interface{}{
				"Namespace":  "secure",
				"Event name": "read",
				"ID":         socket.ID,
				"Input":      message,
			})

			var incomingMessage interface{}

			err = json.Unmarshal(message, &incomingMessage)

			if err != nil {
				errorSocketEvent(fmtLogger, "secure", "unmarshal", socket.ID, err)
			}

			debugSocketEvent(fmtLogger, map[string]interface{}{
				"Namespace":  "secure",
				"Event name": "unmarshal",
				"ID":         socket.ID,
				"Input":      incomingMessage,
			})

			incomingMessageMap := incomingMessage.(map[string]interface{})
			eventName := incomingMessageMap["name"].(string)

			switch eventName {
			case "rsa:getServerKeys":
				getServerKeys("secure", eventName, socket)
			case "rsa:setClientKeys":
				setClientsKeys(incomingMessageMap, "secure", eventName, socket)
			}
		}
	}()

	return nil
}

func (socket *SecureSocket) Send(eventName OutEventName, data string) error {
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
