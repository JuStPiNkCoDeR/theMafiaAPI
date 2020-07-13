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
	"../lib"
	"../logger"
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

// Return complete error message
func createErrorMessage(context context.Context, message string, options *options.ClientOptions) string {
	contextBytes, contextErr := json.Marshal(context)
	ctx := "Unknown"

	if contextErr != nil {
		ctx = string(contextBytes)
	}

	optionBytes, optionErr := json.Marshal(options)
	opt := "Unknown"

	if optionErr != nil {
		opt = string(optionBytes)
	}

	return fmt.Sprintf("%s occurs for database connection with\nContext: %s\nOptions: %s", message, ctx, opt)
}

// Try to establish a connection to the database and save current Client.
func (d *Database) Connect() error {
	client, err := mongo.Connect(d.Context, d.Options)

	if err != nil {
		return &lib.StackError{
			ParentError: err,
			Message:     createErrorMessage(d.Context, "Error on connecting to database", d.Options),
		}
	}

	d.Client = client

	d.Logger.Log(logger.Debug, "Connected to MongoDB!!!", "Database")
	return nil
}

// Send light request to the database to check connection.
func (d *Database) Ping() error {
	err := d.Client.Ping(d.Context, nil)

	if err == nil {
		d.Logger.Log(logger.Debug, "Connection successfully pinged", "Database")

		return nil
	}

	return &lib.StackError{
		ParentError: err,
		Message:     "Error on ping database",
	}
}

// Close the connection
func (d *Database) Close() {
	err := d.Client.Disconnect(d.Context)

	if err == nil {
		d.Logger.Log(logger.Debug, "Connection successfully closed", "Database")
	}
}

// Current connection handles given database
func (d *Database) SelectDatabase(name string, options ...*options.DatabaseOptions) {
	d.CurrentDatabase = d.Client.Database(name, options...)

	d.Logger.Log(logger.Debug, name+" database selected", "Database")
}

// Add the collection to the map of current connection
func (d *Database) AddCollection(name string, options ...*options.CollectionOptions) error {
	if d.CurrentDatabase == nil {
		return &lib.StackError{
			Message: createErrorMessage(d.Context, "You should select database first", d.Options),
		}
	}

	if d.collections == nil {
		d.collections = map[string]*mongo.Collection{}
	}

	d.collections[name] = d.CurrentDatabase.Collection(name, options...)

	d.Logger.Log(logger.Debug, name+" collection added to "+d.CurrentDatabase.Name(), "Database")

	return nil
}

// Return collection from map with specified name
func (d *Database) GetCollection(name string) (collection *mongo.Collection, err error) {
	if collection, ok := d.collections[name]; ok {
		return collection, nil
	}

	return nil, &lib.StackError{
		Message: createErrorMessage(d.Context, "Error: Can't get collection with name "+name, d.Options),
	}
}

// Insert given documents to the specified collection
func (d *Database) Insert(collectionName string, documents []interface{}, key string) error {
	coll, err := d.GetCollection(collectionName)

	if err != nil {
		return &lib.StackError{
			ParentError: err,
			Message:     "Error on getting database collection",
		}
	}

	if len(documents) == 1 {
		d.Logger.Log(
			logger.Debug,
			fmt.Sprintf(
				"Try to insert one document to '%s' collection",
				collectionName,
			),
			key,
		)

		_, err := coll.InsertOne(d.Context, documents[0])

		if err != nil {
			return &lib.StackError{
				ParentError: err,
				Message:     "Error on inserting to database",
			}
		}
	} else if len(documents) == 0 {
		return &lib.StackError{
			Message: "No documents to insert",
		}
	} else {
		d.Logger.Log(
			logger.Debug,
			fmt.Sprintf(
				"Try to insert many documents to '%s' collection",
				collectionName,
			),
			key,
		)

		_, err := coll.InsertMany(d.Context, documents)

		if err != nil {
			return &lib.StackError{
				ParentError: err,
				Message:     "Error on inserting to database",
			}
		}
	}

	return nil
}
