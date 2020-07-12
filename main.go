package main

import (
	"./database"
	"./logger"
	"./utils"
	"time"

	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	port  = 8001
	dbURI = "mongodb://godfather:shawUorResPikT@127.0.0.1:2345/mafia"
)

var (
	fmtLogger     = &logger.MafiaLogger{IsEnabled: true}
	ctx, cancel   = context.WithTimeout(context.Background(), 10*time.Second)
	db            = &database.Database{Logger: fmtLogger, Context: ctx, Options: options.Client().ApplyURI(dbURI)}
	secureClients = make(map[*websocket.Conn]*utils.SecureSocket)
	upgrades      = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

// Handlers
func registrationHandler(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprintf(w, "Reg url")
}

func wsSecureHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrades.Upgrade(w, r, nil)

	if err != nil {
		fmtLogger.Log(logger.Error, fmt.Sprintf("Error on setup of ws connection\nError:\n\t%s", err))
	}

	secureSocket := &utils.SecureSocket{Socket: &utils.Socket{Database: db, Client: ws}}

	err = secureSocket.Init()

	if err != nil {
		fmtLogger.Log(logger.Error, fmt.Sprintf("Error on initializing secure socket\nError:\n\t%s", err))
	}

	secureClients[ws] = secureSocket
}

func main() {
	router := mux.NewRouter()

	// Setup databases
	if err := db.Connect(); err != nil {
		fmtLogger.Log(logger.Error, fmt.Sprintf("Cant establish connection with database(Mongo)\nError:\n\t%s", err.Error()))
		os.Exit(1)
	}
	defer cancel()
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmtLogger.Log(logger.Error, fmt.Sprintf("Error:\n\t%s", err.Error()))
		os.Exit(1)
	}

	db.SelectDatabase("mafia")

	// Adding collections
	if err := db.AddCollection("profiles"); err != nil {
		fmtLogger.Log(logger.Error, fmt.Sprintf("Error on adding database collection\nEroor:\n\t%s", err.Error()))
	}

	fmtLogger.Log(logger.Info, "Connection to the MongoDB established successfully")

	// Setup sockets
	router.HandleFunc("/ws/secure", wsSecureHandler)
	fmtLogger.Log(logger.Info, "WebSocket server setup successfully")

	// Setup routes
	router.HandleFunc("/reg", registrationHandler)
	fmtLogger.Log(logger.Info, "Routes setup successfully")

	fmtLogger.Log(logger.Info, fmt.Sprintf("The application is served on %d port\n", port))
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), router)

	if err != nil {
		fmtLogger.Log(logger.Error, fmt.Sprintf("The program terminated\nError:\n\t%s", err))
	}

	os.Exit(1)
}
