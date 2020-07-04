// Copyright sasha.los.0148@gmail.com
// All Rights have been taken by Mafia :)

// Socket types
package utils

import (
	"../lib"
	"../logger"
	"crypto/x509"
	"encoding/json"
	pem2 "encoding/pem"
	"fmt"
	"github.com/gorilla/websocket"
)

var fmtLogger = &logger.MafiaLogger{IsEnabled: true}

// Log functions
func triggerSocketEvent(output logger.Logger, nsp string, eventName string, ID string) {
	output.Log(logger.Debug, fmt.Sprintf("Triggered socket event\n\tNamespace: %s\n\tEvent name: %s\n\tID: %s\n", nsp, eventName, ID))
}

func successfulSocketEvent(output logger.Logger, nsp string, eventName string, ID string) {
	output.Log(logger.Debug, fmt.Sprintf("Socket event passed OK\n\tNamespace: %s\n\tEvent name: %s\n\tID: %s\n", nsp, eventName, ID))
}

func errorSocketEvent(output logger.Logger, nsp string, eventName string, ID string, err error) {
	output.Log(logger.Error, fmt.Sprintf("Error on socket event\n\tNamespace: %s\n\tEvent name: %s\n\tID: %s\n\tError:\n\t\t%s\n", nsp, eventName, ID, err.Error()))
}

type EventName string

const (
	rsaSendKey    EventName = "rsa:publicKey"
	rsaGetOAEPKey EventName = "rsa:sendKeyOAEP"
)

// Socket response struct
type Response struct {
	Name EventName `json:"name"`
	Data string    `json:"data"`
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

			var incomingMessage interface{}

			err = json.Unmarshal(message, &incomingMessage)

			if err != nil {
				errorSocketEvent(fmtLogger, "secure", "unmarshal", socket.Client.RemoteAddr().String(), err)
			}

			incomingMessageMap := incomingMessage.(map[string]interface{})
			eventName := incomingMessageMap["name"].(EventName)

			switch eventName {
			case "rsa:getKey":
				triggerSocketEvent(fmtLogger, "secure", "getKey", socket.Client.RemoteAddr().String())

				if err = socket.SendPublicKey(); err != nil {
					errorSocketEvent(fmtLogger, "secure", "getKey", socket.Client.RemoteAddr().String(), err)
					return
				}

				successfulSocketEvent(fmtLogger, "secure", "getKey", socket.Client.RemoteAddr().String())
			}
		}
	}()

	return nil
}

func (socket *SecureSocket) Send(eventName EventName, data interface{}) error {
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
func (socket *SecureSocket) SendEncryptedMessage(eventName EventName, data string) error {
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
func (socket *SecureSocket) SendPublicKey() error {
	pem := pem2.EncodeToMemory(&pem2.Block{
		Type:  "PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(socket.r.OwnPublicKey),
	})

	err := socket.Send(rsaSendKey, string(pem))

	if err != nil {
		return &lib.StackError{
			ParentError: err,
			Message:     "Error on writing to socket",
		}
	}

	return nil
}
