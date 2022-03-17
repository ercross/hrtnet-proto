package model

import "time"

type Notification struct {
	Message string `json:"message"`

	// Read specifies if the message has already been read by the receiver
	IsRead bool `json:"is_read"`

	Sent time.Time `json:"sent"`
}
