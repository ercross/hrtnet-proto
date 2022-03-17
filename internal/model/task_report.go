package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type TasksReport struct {
	ID               primitive.ObjectID `json:"id" bson:"id"`
	TelegramUsername string             `json:"telegram_username" bson:"telegram_username"`
	TwitterUsername  string             `json:"twitter_username" bson:"twitter_username"`
	TweetLink        string             `json:"tweet_link" bson:"tweet_link"`
	YoutubeUsername  string             `json:"youtube_username" bson:"youtube_username"`
	WalletAddress    string             `json:"wallet_address" bson:"wallet_address"`

	// EmailAddress can be empty if WalletAddress is provided
	EmailAddress string    `json:"email_address,omitempty" bson:"email_address,omitempty"`
	UserID       string    `json:"user_id" bson:"user_id"`
	SubmittedOn  time.Time `json:"submitted_on" bson:"submitted_on"`
}
