package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	Context context.Context
	Options *options.ClientOptions
	Client  *mongo.Client
}

func (d *Database) Connect() error {
	client, err := mongo.Connect(d.Context, d.Options)

	if err != nil {
		return err
	}

	d.Client = client

	fmt.Println("Connected to MongoDB!!!")
	return nil
}

func (d *Database) Ping() (err error) {
	err = d.Client.Ping(d.Context, nil)

	if err == nil {
		fmt.Println("Connection successfully pinged")
	}

	return
}

func (d *Database) Close() (err error) {
	err = d.Client.Disconnect(d.Context)

	if err == nil {
		fmt.Println("Connection successfully closed")
	}

	return
}