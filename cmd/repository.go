package main

import "github.com/Hrtnet/social-activities/internal/model"

type Repository interface {
	Validator

	// Disconnect the from repo source in case of any fatal event.
	Disconnect() error

	// FetchRandomQRCode returns a string that can be embedded
	// in a QR Code image
	FetchRandomQRCode() (string, error)
	FetchUserID() (string, error)
	FetchAllTaskReports() (*[]model.TasksReport, error)
	FetchTaskReportByUserID(userId string) (*model.TasksReport, error)
	CreateTaskReport(report *model.TasksReport) error
	FetchAllTasks() ([]string, error)

	// InsertMultipleDrugs assigns tracking code (i.e rfid text or alphanum
	// code depending on validationOption) to each drug and inserts same
	// into the repository
	InsertMultipleDrugs(*[]model.Drug, model.ValidationOption) error

	// InsertDrug assigns tracking code (i.e rfid text or alphanum
	// code depending on validationOption) to drug and inserts same
	// into the repository
	InsertDrug(model.Drug, model.ValidationOption) error

	// FetchDrugByBatchNumber fetches the drug bearing this batchNumber
	// and manufactured by this manufacturer.
	// Note that batch numbers are often used for internal tracking
	// by drug manufacturers
	FetchDrugByBatchNumber(batchNumber, manufacturer string) (*model.Drug, error)

	// FetchWalletAddress fetches the wallet address forUserId.
	// Returns db.ErrUserNotFound if user_id is not found in repo
	FetchWalletAddress(forUserId string) (string, error)

	SubmitIncidenceReport(report *model.IncidenceReport) error
}

type Validator interface {
	ValidateQrText(value string) (*model.Drug, error)
	ValidateShortCode(value string) (*model.Drug, error)
	ValidateRFIDText(value string) (*model.Drug, error)
}
