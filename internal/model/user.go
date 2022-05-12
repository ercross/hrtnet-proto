package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {

	// database assigned id
	ID primitive.ObjectID `json:"id" bson:"_id"`

	// system assigned uid
	UID                   string    `json:"user_id" bson:"uid" validate:"required"`
	WalletAddress         string    `json:"wallet_addr" bson:"walletAddr"`
	Email                 string    `json:"email" bson:"email"`
	DateOfBirth           time.Time `json:"dob" bson:"dob"`
	PushNotificationToken string    `json:"push_notification_token" bson:"pushNotificationToken"`
}

// ToMap with bson tag equivalent keys, excluding entry for UID.
// and any other field that has nil/zero value at conversion time
func (u *User) ToMap() map[string]interface{} {
	m := make(map[string]interface{})
	if u.WalletAddress != "" {
		m["walletAddr"] = u.WalletAddress
	}
	if u.Email != "" {
		m["email"] = u.Email
	}
	if !u.DateOfBirth.IsZero() {
		m["dob"] = u.DateOfBirth
	}
	if u.PushNotificationToken != "" {
		m["pushNotificationToken"] = u.PushNotificationToken
	}
	return m
}
