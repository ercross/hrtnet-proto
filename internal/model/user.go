package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {

	// database assigned id
	ID primitive.ObjectID `json:"id" bson:"_id"`

	// system assigned uid
	UID           string `json:"uid" bson:"uid"`
	WalletAddress string `json:"wallet_address" bson:"walletAddr"`
	Email         string `json:"email" bson:"email"`
}
