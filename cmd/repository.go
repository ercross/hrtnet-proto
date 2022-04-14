package main

import "github.com/Hrtnet/social-activities/internal/model"

type Repository interface {
	Validator

	// Disconnect the from repo source in case of any fatal event.
	Disconnect() error

	// FetchRandomQRCode returns a string that can be embedded
	// in a QR Code image
	FetchRandomQRCode() (string, error)
	GenerateNewUserID() (string, error)

	FetchAllAirdropSubmissions() (*[]model.AirdropSubmission, error)

	// FetchAirdropSubmissionByUserID fetches the airdrop submission
	// submitted by userId.
	// Returns db.ErrNoSubmissionFound if no airdrop submission was found.
	FetchAirdropSubmissionByUserID(userId string) (*model.AirdropSubmission, error)

	// InsertAirdropSubmission inserts a airdrop submission document into the database.
	// Note that InsertAirdropSubmission does not check submission.UserID is valid
	InsertAirdropSubmission(submission *model.AirdropSubmission) error

	// IsValidUser checks if id exists in repo.
	// Returns db.ErrUserNotFound if not found, db error otherwise
	IsValidUser(id string) error

	// InsertMultipleDrugs assigns tracking code (i.e rfid text or alphanum
	// code depending on validationOption) to each drug and inserts same
	// into the repository
	InsertMultipleDrugs(*[]model.DBDrug, model.ValidationOption) error

	// FetchUserInfo fetches the model.User.
	// Returns db.ErrUserNotFound if uid is not found in repo
	FetchUserInfo(uid string) (*model.User, error)

	SubmitIncidenceReport(report *model.IncidenceReport) error
}

type NotificationRepo interface {
	SaveNotification(notification *model.Notification) error

	// ReadNotification deletes notification identified by notificationId.
	// If notificationId isn't found in repo, ReadNotification safely returns.
	// Clients can persist notifications that has already been read by user.
	ReadNotification(notificationId string) error

	FetchAllUnreadNotifications(forUserId string) (*[]model.Notification, error)
}

type Validator interface {

	// ValidateQrText validates the text value read from
	// the qr reader.
	// If not found, return db.ErrDrugNotFound
	ValidateQrText(value string) (*model.Drug, error)

	// ValidateShortCode validates short code.
	// If not found, return db.ErrDrugNotFound
	ValidateShortCode(value string) (*model.Drug, error)

	// ValidateRFIDText validates the text value read from
	// the RFID tag.
	// If not found, return db.ErrDrugNotFound
	ValidateRFIDText(value string) (*model.Drug, error)
}
