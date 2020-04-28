package main

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"theMafia/database"
)

func main()  {
	db := database.Database{Context: context.TODO(), Options: options.Client().ApplyURI("mongodb://localhost:2345")}
	err := db.Connect()

	if err != nil {
		log.Fatal("Failed to establish database connection", err)
	}
}
