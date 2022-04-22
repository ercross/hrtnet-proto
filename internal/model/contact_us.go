package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type ContactUs struct {
	ID      primitive.ObjectID `json:"id" bson:"_id"`
	Message string             `json:"message" bson:"message" validate:"required"`
	Title   string             `json:"message" bson:"message" validate:"required"`
	Email   string             `json:"email" bson:"email" validate:"required"`
}
