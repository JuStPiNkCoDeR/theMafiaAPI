package main

import (
	"mafia/sockets/lib"

	"encoding/base64"
	"encoding/json"

	"github.com/gorilla/websocket"
)

// Incoming message for "setClientKeys" event
type SetClientKeysRequest struct {
	oaep string `socket:"require"`
	pss  string `socket:"require"`
}

// Incoming message for "signUp" event
type SignUpRequest struct {
	email    string `socket:"require,rsa,email"`
	password string `socket:"require,rsa"`
}

// Socket response struct
type Response struct {
	Name OutEventName `json:"name"`
	Data string       `json:"data"`
}

// Interface for communication between client and server
type Socket interface {
	Init() error
	Send() error
}

// Simple socket struct
type SimpleSocket struct {
	Client    *websocket.Conn
	Parser    Parser
	ClientID  string
	Namespace string
}

// Setup socket variables
func (socket *SimpleSocket) Init() error {
	socket.ClientID = socket.Client.RemoteAddr().String()
	socket.Parser = &SocketParser{}

	return nil
}

// Send text message
func (socket *SimpleSocket) Send(eventName OutEventName, data string) error {
	response := &Response{
		Name: eventName,
		Data: data,
	}

	resJSON, err := json.Marshal(response)
	if err != nil {
		return lib.Wrap(err, "Error on marshal response object to json")
	}

	err = socket.Client.WriteMessage(websocket.TextMessage, resJSON)
	if err != nil {
		return lib.Wrap(err, "Error on writing to sockets")
	}

	return nil
}

// Simple socket with RSA implementation
type SecureSocket struct {
	*SimpleSocket
	rsa *RSA
}

// Setup secure sockets' variables
func (socket *SecureSocket) Init() error {
	if err := socket.SimpleSocket.Init(); err != nil {
		return lib.Wrap(err, "Error on initializing of secure sockets")
	}

	socket.Parser = &SocketParser{
		rsa: socket.rsa,
	}
	socket.rsa = &RSA{}
	socket.Namespace = "secure"

	if err := socket.rsa.Init(); err != nil {
		return lib.Wrap(err, "Error on initializing of secure sockets")
	}

	return nil
}

// Accepts handshake requests, profile requests
func (socket *SecureSocket) handle() {
	done := make(chan string)

	go func() {
		defer close(done)

		for {
			var (
				incomingMessage    interface{}
				incomingMessageMap DataObject
				input              DataObject
				eventName          string
				requestID          string
			)

			fmtLogger.LogKey = "NONE"
			// Read new message from client
			_, message, err := socket.Client.ReadMessage()
			if err != nil {
				// Close event handler
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
					errorSocketEvent(fmtLogger, "read", err, socket.SimpleSocket)
				} else {
					debugSocketEvent(fmtLogger, "Socket close normally")
				}
				break
			}

			debugSocketEvent(fmtLogger, map[string]interface{}{
				"Namespace":  socket.Namespace,
				"Event name": "read",
				"ID":         socket.ClientID,
				"Input":      message,
			})

			// Parse JSON string from message
			err = json.Unmarshal(message, &incomingMessage)
			if err != nil {
				errorSocketEvent(fmtLogger, "unmarshal", err, socket.SimpleSocket)
			}

			debugSocketEvent(fmtLogger, map[string]interface{}{
				"Namespace":  socket.Namespace,
				"Event name": "unmarshal",
				"ID":         socket.ClientID,
				"Input":      incomingMessage,
			})

			// Get event name and request ID
			incomingMessageMap = incomingMessage.(map[string]interface{})
			eventName = incomingMessageMap["name"].(string)
			requestID = incomingMessageMap["reqID"].(string)
			input = incomingMessageMap["data"].(DataObject)

			fmtLogger.LogKey = requestID

			switch eventName {
			// Client wants to get servers RSA public keys
			case "rsa:getServerKeys":
				getServerKeys(eventName, socket)
			// Client sends own RSA public keys
			case "rsa:setClientKeys":
				data, err := socket.Parser.Parse(input, &SetClientKeysRequest{})
				if err != nil {
					errorSocketEvent(fmtLogger, eventName, lib.Wrap(err, "Error on parsing input data"), socket.SimpleSocket)

					if err := socket.Send(rsaAcceptClientKeys, "NO"); err != nil {
						errorSocketEvent(fmtLogger, eventName, lib.Wrap(err, "Error on sending message"), socket.SimpleSocket)
					}
				}

				debugSocketEvent(fmtLogger, map[string]interface{}{
					"event":   eventName,
					"message": "Parsed data object",
					"parsed":  data,
				})

				if isAccepted := setClientsKeys(data, eventName, socket); isAccepted {
					if err := socket.Send(rsaAcceptClientKeys, "YES"); err != nil {
						errorSocketEvent(fmtLogger, eventName, lib.Wrap(err, "Error on sending message"), socket.SimpleSocket)
					}
				} else {
					if err := socket.Send(rsaAcceptClientKeys, "NO"); err != nil {
						errorSocketEvent(fmtLogger, eventName, lib.Wrap(err, "Error on sending message"), socket.SimpleSocket)
					}
				}
			// Client signing up
			case "rsa:signUp":
				// gRPC request to profile service
				data, err := socket.Parser.Parse(input, &SignUpRequest{})
				if err != nil {
					errorSocketEvent(fmtLogger, eventName, lib.Wrap(err, "Error on parsing input data"), socket.SimpleSocket)
				}

				debugSocketEvent(fmtLogger, map[string]interface{}{
					"event":   eventName,
					"message": "Parsed data object",
					"parsed":  data,
				})
			}
		}
	}()
}

// Send encrypted single message
func (socket *SecureSocket) SendEncryptedMessage(eventName OutEventName, data string) error {
	encrypted, err := socket.rsa.Encode(data, "")

	if err == nil {
		if err := socket.Send(eventName, base64.StdEncoding.EncodeToString(encrypted)); err != nil {
			return err
		}

		return nil
	}

	return &lib.StackError{
		ParentError: err,
		Message:     "Error on sockets encryption",
	}
}

// Send RSA public key as JSON string
func (socket *SecureSocket) SendPublicKeys() error {
	pemOAEP, pemPSS, err := socket.rsa.ExportKeys()
	if err != nil {
		return lib.Wrap(err, "Can't export the own public keys")
	}

	debugSocketEvent(fmtLogger, map[string]interface{}{
		"Event name": "Encode OAEP key to pem",
		"PEM":        pemOAEP,
	})

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
			Message:     "Error on writing to sockets",
		}
	}

	return nil
}
