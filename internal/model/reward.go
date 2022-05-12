package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Reward struct {
	ID        primitive.ObjectID `json:"id" bson:"id"`
	Points    int                `json:"points" bson:"points"`
	HrtTokens int                `json:"hrt_tokens" bson:"hrt_tokens"`
}
