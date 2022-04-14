package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type AirdropSubmission struct {
	ID               primitive.ObjectID `json:"id" bson:"_id"`
	TelegramUsername string             `json:"telegram_username" bson:"telegramUsername" validate:"required"`
	TwitterUsername  string             `json:"twitter_username" bson:"twitterUsername" validate:"required"`
	TweetLink        string             `json:"tweet_link" bson:"tweetLink" validate:"required"`
	YoutubeUsername  string             `json:"youtube_username" bson:"youtubeUsername"`

	// WalletAddress can be empty if EmailAddress is empty
	WalletAddress string `json:"wallet_address" bson:"wallet,omitempty"`

	// EmailAddress can be empty if WalletAddress is provided
	EmailAddress string    `json:"email_address,omitempty" bson:"email,omitempty"`
	UserID       string    `json:"user_id" bson:"uid" validate:"required"`
	SubmittedOn  time.Time `json:"submitted_on" bson:"submittedOn" validate:"required"`
}
