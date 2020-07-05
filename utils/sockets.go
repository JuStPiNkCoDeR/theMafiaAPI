// Copyright sasha.los.0148@gmail.com
// All Rights have been taken by Mafia :)

// Socket types
package utils

import (
	"../lib"
	"../logger"
	"crypto/x509"
	"encoding/asn1"
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

// Socket response struct
type Response struct {
	Name OutEventName `json:"name"`
	Data string       `json:"data"`
}

// Socket that use RSA encryption
type SecureSocket struct {
	r      *RSA            // RSA object to encode/decode
	Client *websocket.Conn // Connection instance
}

// Init socket
func (socket *SecureSocket) Init() error {
	socket.r = &RSA{}

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
				errorSocketEvent(fmtLogger, "secure", "read", socket.Client.RemoteAddr().String(), err)
				return
			}

			debugSocketEvent(fmtLogger, map[string]interface{}{
				"Namespace":  "secure",
				"Event name": "read",
				"ID":         socket.Client.RemoteAddr().String(),
				"Input":      message,
			})

			var incomingMessage interface{}

			err = json.Unmarshal(message, &incomingMessage)

			if err != nil {
				errorSocketEvent(fmtLogger, "secure", "unmarshal", socket.Client.RemoteAddr().String(), err)
			}

			debugSocketEvent(fmtLogger, map[string]interface{}{
				"Namespace":  "secure",
				"Event name": "unmarshal",
				"ID":         socket.Client.RemoteAddr().String(),
				"Input":      incomingMessage,
			})

			incomingMessageMap := incomingMessage.(map[string]interface{})
			eventName := incomingMessageMap["name"]

			switch eventName {
			case "rsa:getServerKeys":
				triggerSocketEvent(fmtLogger, "secure", "getKey", socket.Client.RemoteAddr().String())

				if err = socket.SendPublicKeys(); err != nil {
					errorSocketEvent(fmtLogger, "secure", "getKey", socket.Client.RemoteAddr().String(), err)
					return
				}

				successfulSocketEvent(fmtLogger, "secure", "getKey", socket.Client.RemoteAddr().String())
			case "rsa:setClientKeys":

			}
		}
	}()

	return nil
}

func (socket *SecureSocket) Send(eventName OutEventName, data interface{}) error {
	jsonData, err := json.Marshal(data)

	if err != nil {
		return &lib.StackError{
			ParentError: err,
			Message:     "Error on marshal to json",
		}
	}

	response := &Response{
		Name: eventName,
		Data: string(jsonData),
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
		if err := socket.Send(eventName, encrypted); err != nil {
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

	bytesPSS, err := asn1.Marshal(*socket.r.OwnPublicKeyPSS)

	if err != nil {
		return &lib.StackError{
			ParentError: err,
			Message:     "Error on ASN marshal for PSS public key",
		}
	}

	pemPSS := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: bytesPSS,
	})

	if pemPSS == nil {
		return &lib.StackError{
			Message: "Error on encoding public key(PSS)",
		}
	}

	data := map[string]interface{}{
		"OAEP": pemOAEP,
		"PSS":  pemPSS,
	}

	err = socket.Send(rsaSendServerKeys, data)

	if err != nil {
		return &lib.StackError{
			ParentError: err,
			Message:     "Error on writing to socket",
		}
	}

	return nil
}
