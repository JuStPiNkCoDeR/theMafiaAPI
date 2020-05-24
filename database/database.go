// Copyright sasha.los.0148@gmail.com
// All Rights have been taken by Mafia :)

// Database connection manager.
// We are using MongoDB for storing data.
//
// For more information see: https://www.mongodb.com/.
//
// MongoDB Driver example for Go: https://www.mongodb.com/blog/post/quick-start-golang--mongodb--modeling-documents-with-go-data-structures.
package database

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"theMafia/logger"
)

// Manager structure.
type Database struct {
	Logger          logger.Logger          // Database logger
	Context         context.Context        // Context of the current connection.
	Options         *options.ClientOptions // The connection options.
	Client          *mongo.Client          // The connection.
	CurrentDatabase *mongo.Database        // database selected via SelectDatabase func
	collections     map[string]*mongo.Collection
}

// Database error
type Error struct {
	ParentError error                  // Error which has been thrown
	Message     string                 // String contains short information about error
	Context     context.Context        // Context of connection
	Options     *options.ClientOptions // Options of connection
}

// Return complete error message
func (e *Error) Error() string {
	contextBytes, contextErr := json.Marshal(e.Context)
	ctx := "Unknown"

	if contextErr != nil {
		ctx = string(contextBytes)
	}

	optionBytes, optionErr := json.Marshal(e.Options)
	opt := "Unknown"

	if optionErr != nil {
		opt = string(optionBytes)
	}

	return fmt.Sprintf("%s occurs for database connection with\nContext: %s\nOptions: %s", e.Message, ctx, opt)
}

// Try to establish a connection to the database and save current Client.
func (d *Database) Connect() error {
	client, err := mongo.Connect(d.Context, d.Options)

	if err != nil {
		return err
	}

	d.Client = client

	d.Logger.Log("debug", "Connected to MongoDB!!!")
	return nil
}

// Send light request to the database to check connection.
func (d *Database) Ping() (err error) {
	err = d.Client.Ping(d.Context, nil)

	if err == nil {
		d.Logger.Log("debug", "Connection successfully pinged")
	}

	return
}

// Close the connection
func (d *Database) Close() {
	err := d.Client.Disconnect(d.Context)

	if err == nil {
		d.Logger.Log("debug", "Connection successfully closed")
	}
}

// Current connection handles given database
func (d *Database) SelectDatabase(name string, options ...*options.DatabaseOptions) {
	d.CurrentDatabase = d.Client.Database(name, options...)

	d.Logger.Log("debug", name+" database selected")
}

// Add the collection to the map of current connection
func (d *Database) AddCollection(name string, options ...*options.CollectionOptions) error {
	if d.CurrentDatabase == nil {
		return &Error{
			Message: "You should select database first",
			Context: d.Context,
			Options: d.Options,
		}
	}

	if d.collections == nil {
		d.collections = map[string]*mongo.Collection{}
	}

	d.collections[name] = d.CurrentDatabase.Collection(name, options...)

	d.Logger.Log("debug", name+" collection added to "+d.CurrentDatabase.Name())

	return nil
}

// Return collection from map with specified name
func (d *Database) GetCollection(name string) (collection *mongo.Collection, err error) {
	if collection, ok := d.collections[name]; ok {
		return collection, nil
	} else {
		return collection, &Error{
			Message: "Error: Can't get collection with name " + name,
			Context: d.Context,
			Options: d.Options,
		}
	}
}
