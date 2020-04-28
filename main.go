package main

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo/options"
	"theMafia/database"
)

func main()  {
	db := database.Database{Context: context.TODO(), Options: options.Client().ApplyURI("mongodb://localhost:2345")}
	db.Connect()
}
