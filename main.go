package main

import (
	"./logger"
	"./utils"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const port = 8001

var fmtLogger = &logger.MafiaLogger{IsEnabled: true}
var secureClients = make(map[*websocket.Conn]*utils.SecureSocket)
var upgrades = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handlers
func registrationHandler(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprintf(w, "Reg url")
}

func wsSecureHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrades.Upgrade(w, r, nil)

	if err != nil {
		fmtLogger.Log(logger.Error, fmt.Sprintf("Error on setup of ws connection\nError:\n\t%s", err))
	}

	secureSocket := &utils.SecureSocket{Client: ws}

	err = secureSocket.Init()

	if err != nil {
		fmtLogger.Log(logger.Error, fmt.Sprintf("Error on initializing secure socket\nError:\n\t%s", err))
	}

	secureClients[ws] = secureSocket
}

func main() {
	router := mux.NewRouter()

	// Setup sockets
	router.HandleFunc("/ws/secure", wsSecureHandler)
	fmtLogger.Log(logger.Info, "WebSocket server setup successfully")

	// Setup routes
	router.HandleFunc("/reg", registrationHandler)
	fmtLogger.Log(logger.Info, "Routes setup successfully")

	fmtLogger.Log(logger.Info, fmt.Sprintf("The application is served on %d port\n", port))
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), router)

	fmtLogger.Log(logger.Error, fmt.Sprintf("The program terminated\nError:\n\t%s", err))
}
