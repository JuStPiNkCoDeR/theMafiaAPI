package main

import (
	"mafia/sockets/lib"

	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type OutEventName string

const (
	rsaSendServerKeys   OutEventName = "rsa:serverKeys"
	rsaAcceptClientKeys OutEventName = "rsa:acceptClientKeys"
	rsaSignUp           OutEventName = "rsa:signUp"
)

var (
	fmtLogger     = &lib.MafiaLogger{IsEnabled: true}
	secureClients = make(map[*websocket.Conn]*SecureSocket)
	upgrades      = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func secureSocketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrades.Upgrade(w, r, nil)

	if err != nil {
		fmtLogger.Log(lib.Error, fmt.Sprintf("Error on setup of ws connection\nError:\n\t%s", err))
	}

	secureSocket := &SecureSocket{SimpleSocket: &SimpleSocket{Client: ws}}

	err = secureSocket.Init()
	if err != nil {
		fmtLogger.Log(lib.Error, fmt.Sprintf("Error on initializing secure sockets\nError:\n\t%s", err))
	}

	secureClients[ws] = secureSocket
}

func main() {
	fmtLogger.LogKey = "DEPLOY"

	fmt.Print("Hello from Service!!!")
}
