package unit

import (
	"../../database"
	"../../logger"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
)

const dbURI = "mongodb://test:test@127.0.0.1:2345/test"

var mafiaLogs = &logger.MafiaLogger{IsEnabled: false}
var testDB = database.Database{Logger: mafiaLogs, Context: context.TODO(), Options: options.Client().ApplyURI(dbURI)}

func TestDatabase_Connect(t *testing.T) {
	err := testDB.Connect()

	if err != nil {
		t.Fatal("Got error on connection", err)
	}
}

func TestDatabase_Ping(t *testing.T) {
	err := testDB.Ping()

	if err != nil {
		t.Fatal("Got error on ping connection", err)
	}
}

func TestDatabase_SelectDatabase(t *testing.T) {
	testDB.SelectDatabase("mafia")

	var databaseName = testDB.CurrentDatabase.Name()
	if databaseName != "mafia" {
		t.Fatal("Got different database name\nExpected: mafia\nGot: ", databaseName)
	}
}

func TestDatabase_AddCollection(t *testing.T) {
	err := testDB.AddCollection("test")

	if err != nil {
		t.Fatal("Something went wrong with adding collection", err)
	}

	if _, ok := testDB.GetCollection("test"); ok != nil {
		t.Fatal("Cant't get collection with name 'test'", ok)
	}
}

func TestDatabase_Close(t *testing.T) {
	testDB.Close()

	err := testDB.Ping()

	if err == nil {
		t.Fatal("Connection has not been closed")
	}
}

func ExampleDatabase_Connect() {
	var db database.Database
	mafiaLogs := &logger.MafiaLogger{IsEnabled: false}

	db = database.Database{Logger: mafiaLogs, Context: context.TODO(), Options: options.Client().ApplyURI("mongodb://localhost:2345")}
	err := db.Connect()

	if err != nil {
		panic("Can't connect to database")
	}
	defer db.Close()
}

func ExampleDatabase_Ping() {
	var db database.Database
	mafiaLogs := &logger.MafiaLogger{IsEnabled: false}

	db = database.Database{Logger: mafiaLogs, Context: context.TODO(), Options: options.Client().ApplyURI("mongodb://localhost:2345")}
	err := db.Connect()

	if err != nil {
		panic("Can't connect to database")
	}
	defer db.Close()

	err = db.Ping()

	if err != nil {
		panic("Can't ping connection")
	}
}

func ExampleDatabase_Close() {
	var db database.Database
	mafiaLogs := &logger.MafiaLogger{IsEnabled: false}

	db = database.Database{Logger: mafiaLogs, Context: context.TODO(), Options: options.Client().ApplyURI("mongodb://localhost:2345")}
	err := db.Connect()

	if err != nil {
		panic("Can't connect to database")
	}
	defer db.Close()
}

func ExampleDatabase_SelectDatabase() {
	var db database.Database
	mafiaLogs := &logger.MafiaLogger{IsEnabled: false}

	db = database.Database{Logger: mafiaLogs, Context: context.TODO(), Options: options.Client().ApplyURI("mongodb://localhost:2345")}
	err := db.Connect()

	if err != nil {
		panic("Can't connect to database")
	}
	defer db.Close()

	db.SelectDatabase("mafia")
}

func ExampleDatabase_AddCollection() {
	var db database.Database
	mafiaLogs := &logger.MafiaLogger{IsEnabled: false}

	db = database.Database{Logger: mafiaLogs, Context: context.TODO(), Options: options.Client().ApplyURI("mongodb://localhost:2345")}
	err := db.Connect()

	if err != nil {
		panic("Can't connect to database")
	}
	defer db.Close()

	db.SelectDatabase("mafia")
	err = db.AddCollection("test")

	if err != nil {
		panic("Can't add a collection")
	}
}

func ExampleDatabase_GetCollection() {
	var db database.Database
	mafiaLogs := &logger.MafiaLogger{IsEnabled: false}

	db = database.Database{Logger: mafiaLogs, Context: context.TODO(), Options: options.Client().ApplyURI("mongodb://localhost:2345")}
	err := db.Connect()

	if err != nil {
		panic("Can't connect to database")
	}
	defer db.Close()

	db.SelectDatabase("mafia")
	err = db.AddCollection("test")

	if err != nil {
		panic("Can't add a collection")
	}

	if collection, ok := db.GetCollection("test"); ok == nil {
		fmt.Println(collection.Name())
	}
	// Output: test
}
