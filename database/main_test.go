package database

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
)

var db = Database{Context: context.TODO(), Options: options.Client().ApplyURI("mongodb://localhost:2345")}

func TestDatabase_Connect(t *testing.T) {
	err := db.Connect()

	if err != nil {
		t.Error("Got error on connection", err)
	}
}

func TestDatabase_Ping(t *testing.T) {
	err := db.Ping()

	if err != nil {
		t.Error("Got error on ping connection", err)
	}
}

func TestDatabase_Close(t *testing.T) {
	err := db.Close()

	if err != nil {
		t.Error("Got error on connection close", err)
	}
}