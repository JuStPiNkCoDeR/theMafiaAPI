package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Profile struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	Password  string             `bson:"password,omitempty"`
	CreatedAt primitive.DateTime `bson:"created_at"`
	DeletedAt primitive.DateTime `bson:"deleted_at,omitempty"`
}
