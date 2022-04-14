package model

import (
	"fmt"
	"github.com/Hrtnet/social-activities/internal/logger"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"time"
)

type Notification struct {
	ID string `json:"id"`

	UserID string `json:"user_id" bson:"uid"`

	Title string `json:"title" bson:"title"`

	Message string `json:"message" bson:"message"`

	// Read specifies if the message has already been read by the receiver
	IsRead bool `json:"is_read" bson:"isRead"`

	Sent time.Time `json:"sent" bson:"sent"`
}

// InsertID inserts ID into Notification.
// A UUID generator is used to avoid possible delays
// that might be experienced if we chose to use a database assigned id.
// InsertID must be invoked on Notification if the instance of Notification was not created
// by any of the Notification-returning methods below
func (notification *Notification) InsertID() {

	defer func() {
		if r := recover(); r != nil {
			logger.Logger.LogWarn(
				"recovering from new UUID string panic",
				"insert notification id",
				errors.New(fmt.Sprintf("%v", r)))
		}
	}()

	// The code below may cause a panic.
	// Check https://pkg.go.dev/github.com/google/uuid#NewString
	notification.ID = uuid.NewString()
}

func NewWelcomeNotification(userId string) *Notification {

	notification := &Notification{
		UserID:  userId,
		Title:   "Welcome to HeartNet",
		Message: fmt.Sprintf("Thanks for signing in with HeartNet. Your UID (user id) is %s", userId),
		IsRead:  false,
		Sent:    time.Now(),
	}
	notification.InsertID()
	return notification
}

func NewWelcomeBackNotification(userId string) *Notification {
	notification := &Notification{
		UserID:  userId,
		Title:   fmt.Sprintf("Welcome Back %s", userId),
		Message: "Welcome back to HeartNet. We are still committed to promoting healthy habits that reduce the rate of hypertension in Africa",
		IsRead:  false,
		Sent:    time.Now(),
	}
	notification.InsertID()
	return notification
}

func NewIncidenceReportNotification(userId string) *Notification {
	notification := &Notification{
		UserID:  userId,
		Title:   "Incidence Report SubmittedOn",
		Message: "We have received your incidence report and our investigative partners will look into the report. Thanks.",
		IsRead:  false,
		Sent:    time.Now(),
	}
	notification.InsertID()
	return notification
}

func NewTaskReportNotification(userId string) *Notification {
	notification := &Notification{
		UserID:  userId,
		Title:   "Participation Recorded",
		Message: "Thank you for participating in our airdrop program. Your submission has been recorded and the rewards will be distributed as at when due",
		IsRead:  false,
		Sent:    time.Now(),
	}
	notification.InsertID()
	return notification
}

// NewValidationNotification generates a validation notification.
func NewValidationNotification(userId, validationResult string) *Notification {
	notification := &Notification{
		UserID:  userId,
		Title:   "Validation Report",
		Message: fmt.Sprintf("You have just conducted a validation through HeartNet DApp and the drug was %s", validationResult),
		IsRead:  false,
		Sent:    time.Now(),
	}
	notification.InsertID()
	return notification
}
