package main

import (
	"encoding/json"
	"fmt"
	"mafia/sockets/lib"
)

// Log functions
func debugSocketEvent(outStream lib.Logger, data interface{}) {
	dataJSON, err := json.Marshal(data)

	if err != nil {
		outStream.Log(lib.Error, fmt.Sprintf("Unable to serialize data for debug\nError:\n\t%s", err))
		return
	}

	outStream.Log(lib.Debug, fmt.Sprintf("Debug sockets event\nData:\n\t%s", string(dataJSON)))
}

func triggerSocketEvent(outStream lib.Logger, eventName string, socket *SimpleSocket) {
	outStream.Log(lib.Debug, fmt.Sprintf("Triggered sockets event\n\tNamespace: %s\n\tEvent name: %s\n\tID: %s", socket.Namespace, eventName, socket.ClientID))
}

func successfulSocketEvent(outStream lib.Logger, eventName string, socket *SimpleSocket) {
	outStream.Log(lib.Debug, fmt.Sprintf("Socket event passed OK\n\tNamespace: %s\n\tEvent name: %s\n\tID: %s", socket.Namespace, eventName, socket.ClientID))
}

func errorSocketEvent(outStream lib.Logger, eventName string, err error, socket *SimpleSocket) {
	outStream.Log(lib.Error, fmt.Sprintf("Error on sockets event\n\tNamespace: %s\n\tEvent name: %s\n\tID: %s\n\tError:\n\t%s", socket.Namespace, eventName, socket.ClientID, err.Error()))
}
