package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Announcement struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`

	// May not be displayed to end user.
	// Used internally as a secondary id for announcement,
	// especially in file system
	Title string `json:"title" bson:"title" validate:"required,max=35"`

	// Link to any image resources.
	// Image should typically be an infographics of
	// the announcement
	ImageUrl string `json:"image_url" bson:"imageUrl" validate:"required"`

	// Text to be used as push notification body, not more than 150 character
	Text string `json:"text" bson:"text" validate:"max=150"`

	// Link to any internet resources
	Url       string    `json:"url" bson:"url" validate:"url"`
	ValidTill time.Time `json:"valid_till" bson:"validTill" validate:"required"`
	CreatedOn time.Time `json:"created_on" bson:"createdOn"`
}
