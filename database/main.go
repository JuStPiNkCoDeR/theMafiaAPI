package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

type Database struct {
	Context context.Context
	Options *options.ClientOptions
	Client  *mongo.Client
}

func (d *Database) Connect() {
	client, err := mongo.Connect(d.Context, d.Options)

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(d.Context, nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!!!")
}