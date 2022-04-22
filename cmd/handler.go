package main

import (
	"fmt"
	"github.com/Hrtnet/social-activities/internal/db"
	"github.com/Hrtnet/social-activities/internal/logger"
	"github.com/Hrtnet/social-activities/internal/model"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/skip2/go-qrcode"
	"net/http"
	"strings"
	"time"
)

// submitContactUsMessage
// Method: POST
// Parameters:
//		email string *required
//		title string *required
//		message string *required
func (app *app) submitContactUsMessage(w http.ResponseWriter, r *http.Request) {
	var msg model.ContactUs
	err := app.readJSON(w, r, &msg)
	if err != nil {
		app.sendBadRequestResponse(w, r, err)
		return
	}

	validate := validator.New()
	if err := validate.Struct(msg); err != nil {
		errs := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			errs[err.Field()] = err.Error()
		}
		app.sendFailedValidationResponse(w, r, errs)
		return
	}

	if err := app.repo.InsertContactUs(&msg); err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}

	app.sendAPIResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message:    "message recorded",
	}, r, nil)
}

// updateEmail
// Method: POST
// Parameters:
// 		email string *required
// 		user_id string *required
func (app *app) updateEmail(w http.ResponseWriter, r *http.Request) {
	type p struct {
		Email string `json:"email" validate:"required,email"`
		UID   string `json:"user_id" validate:"required"`
	}
	var user p
	err := app.readJSON(w, r, &user)
	if err != nil {
		app.sendBadRequestResponse(w, r, err)
		return
	}

	validate := validator.New()
	if err := validate.Struct(user); err != nil {
		errs := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			errs[err.Field()] = err.Error()
		}
		app.sendFailedValidationResponse(w, r, errs)
		return
	}

	if err := app.repo.UpdateUserEmail(user.Email, user.UID); err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}

	app.sendAPIResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message:    "update user successful",
	}, r, nil)
}

func (app *app) updateWalletAddress(w http.ResponseWriter, r *http.Request) {
	type p struct {
		WalletAddr string `json:"wallet_addr" validate:"required"`
		UID        string `json:"user_id" validate:"required"`
	}
	var user p
	err := app.readJSON(w, r, &user)
	if err != nil {
		app.sendBadRequestResponse(w, r, err)
		return
	}

	validate := validator.New()
	if err := validate.Struct(user); err != nil {
		errs := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			errs[err.Field()] = err.Error()
		}
		app.sendFailedValidationResponse(w, r, errs)
		return
	}

	if err := app.repo.UpdateUserWalletAddress(user.WalletAddr, user.UID); err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}

	app.sendAPIResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message:    "update user successful",
	}, r, nil)
}

// serveAllAirdropSubmission serves activities statistics
// METHOD: GET
// Request must contain admin authorization
func (app *app) serveAllAirdropSubmission(w http.ResponseWriter, r *http.Request) {
	reports, err := app.repo.FetchAllAirdropSubmissions()
	if err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}

	app.sendAPIResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message:    "All participants' reports",
	}, r, reports)
	return
}

// serveAirdropSubmission
// returns the task report submitted by user identified by
// user_id in query parameter
// METHOD: GET
// Query parameter: user_id
func (app *app) serveAirdropSubmission(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("user_id")
	report, err := app.repo.FetchAirdropSubmissionByUserID(userId)
	if err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}

	app.sendAPIResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message:    "Welcome to HeartNet",
	}, r, report)
	return
}

// submitAirdropForm
// METHOD: POST
// Accept: application/json
// Request Body fields
// 			telegram_username string *required
//			twitter_username string *required
//			tweet_link string *required
// 			wallet_address string (must be present if email_address is absent
//			email_address string (must be present if wallet_address is absent
// 			user_id string *required
func (app *app) submitAirdropForm(w http.ResponseWriter, r *http.Request) {

	var report model.AirdropSubmission
	err := app.readJSON(w, r, &report)
	if err != nil {
		app.sendBadRequestResponse(w, r, err)
		return
	}

	if errs := validateAirdropSubmission(&report); errs != nil {
		app.sendFailedValidationResponse(w, r, errs)
		return
	}

	// verify that user exist
	user, err := app.repo.FetchUserInfo(report.UserID)
	if err != nil {
		if err == db.ErrUserNotFound || user.UID == "" {
			app.sendBadRequestResponse(w, r, errors.New("UID missing"))
			return
		}
	}

	// verify that user hasn't made any previous submission
	submission, err := app.repo.FetchAirdropSubmissionByUserID(user.UID)
	if err != nil && err != db.ErrNoSubmissionFound {
		app.sendServerErrorResponse(w, r, err)
		return
	}
	if submission != nil {
		app.sendEditConflictResponse(w, r, "airdrop participation already recorded")
		return
	}

	if err = app.repo.InsertAirdropSubmission(&report); err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}

	app.notificationHub.Dispatch(model.NewTaskReportNotification(report.UserID))
	app.sendAPIResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message:    "Your participation has been recorded successfully",
	}, r, nil)
	return
}

// validateQrCode
// Method: POST
// Content-Type application/json

// Request Body fields
// data string *required (the text resulting from QR code scan)
func (app *app) validateQrCode(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Data   string `json:"data"`
		UserID string `json:"user_id"`
	}
	err := app.readJSON(w, r, &in)
	if err != nil {
		app.sendBadRequestResponse(w, r, err)
		return
	}
	if err := app.repo.IsValidUser(in.UserID); err != nil {
		errs := map[string]string{"user_id": "invalid user id"}
		app.sendFailedValidationResponse(w, r, errs)
		return
	}

	drug, err := app.repo.ValidateQrText(in.Data)
	app.processValidation(w, r, drug, in.UserID, err)
}

// validateShortCode
// Method: POST
// Accept application/json
// Request Body
// 		data string *required (the short code)
func (app *app) validateShortCode(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Data   string `json:"data"`
		UserID string `json:"user_id"`
	}
	err := app.readJSON(w, r, &in)
	if err != nil {
		app.sendBadRequestResponse(w, r, err)
		return
	}

	if err := app.repo.IsValidUser(in.UserID); err != nil {
		errs := map[string]string{"user_id": "invalid user id"}
		app.sendFailedValidationResponse(w, r, errs)
		return
	}

	drug, err := app.repo.ValidateShortCode(in.Data)
	app.processValidation(w, r, drug, in.UserID, err)
}

// validateShortCode
// Method: POST
// Accept application/json
// Request Body Fields
// 		data string *required (the string read from the RFID tag)
func (app *app) validateRFIDText(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Data   string `json:"data"`
		UserID string `json:"user_id"`
	}
	err := app.readJSON(w, r, &in)
	if err != nil {
		app.sendBadRequestResponse(w, r, err)
		return
	}

	if err := app.repo.IsValidUser(in.UserID); err != nil {
		errs := map[string]string{"user_id": "invalid user id"}
		app.sendFailedValidationResponse(w, r, errs)
		return
	}

	drug, err := app.repo.ValidateShortCode(in.Data)
	app.processValidation(w, r, drug, in.UserID, err)
}

// serveQrCode serves a single QrCode instance to client
// METHOD: GET
// Content-Type: image/png
// The served QRCode is an image,
func (app *app) serveQrCode(w http.ResponseWriter, r *http.Request) {
	qrCode, err := app.repo.FetchRandomQRCode()
	if err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	raw, err := qrcode.Encode(qrCode, qrcode.Medium, 256)
	if err != nil {
		app.sendServerErrorResponse(w, r, errors.Wrap(err, "error converting qr code to png"))
		return
	}

	w.WriteHeader(200)
	w.Write(raw)
	logger.Logger.LogServe(200, r)

}

// serveStarterPack serves new user with userId and a wallet address
// METHOD: GET
// Content-Type: application/json
func (app *app) serveStarterPack(w http.ResponseWriter, r *http.Request) {
	userId, err := app.repo.GenerateNewUserID()
	if err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}

	app.sendAPIResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message:    "Welcome to HeartNet",
	}, r, map[string]interface{}{
		"user_id": userId,
	})
	app.notificationHub.Dispatch(model.NewWelcomeNotification(userId))
	return
}

func (app *app) checkStatus(w http.ResponseWriter, r *http.Request) {
	app.sendAPIResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message:    "HeartNet prototype is up and running",
	}, r, nil)
}

// serveStarterPack serves returning user with their existing wallet address
// METHOD: GET
// Content-Type: application/json
// Query param: user_id string *required
func (app *app) serveWalletAddress(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("user_id")
	if userId == "" {
		app.sendBadRequestResponse(w, r, errors.New("missing user_id"))
		return
	}

	user, err := app.repo.FetchUserInfo(userId)
	if err == db.ErrUserNotFound {
		app.sendAPIResponse(&responseWriterArgs{
			writer:     w,
			statusCode: 404,
			status:     false,
			message:    "user id not found",
		}, r, nil)
		return
	}
	if err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}
	app.sendAPIResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message:    fmt.Sprintf("%s wallet address", userId),
	}, r, map[string]string{"address": user.WalletAddress})
	app.notificationHub.Dispatch(model.NewWelcomeBackNotification(userId))
}

var errFileTooLarge = errors.New("file too large")

// submitIncidenceReport
// METHOD: POST
// Content-type: multipart/form-data
// Request Body:
//		user_id string *required
//		pharmacy_name string *required
//		description string *required
//		pharmacy_location string *required
// 		evidence_images multipartfile (Content-Type: application/png, Content-Encoding: zip) *required
//		receipt multipartfile (Content-Type file/image)
//
// Since there's currently no cap on the number of evidence_images that can be
// attached to the request, all evidence_images must be zipped (i.e compressed as zip)
func (app *app) submitIncidenceReport(w http.ResponseWriter, r *http.Request) {

	// Max memory::32 MB
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		app.sendBadRequestResponse(w, r, err)
		return
	}

	report, errorType, errs := app.extractIncidenceReport(r, app.config)
	if len(errs) != 0 {
		if errorType == errInternal {
			app.sendServerErrorResponse(w, r, errors.New(errs["error"]))
			return
		}
		app.sendFailedValidationResponse(w, r, errs)
		return
	}

	if err := app.repo.SubmitIncidenceReport(report); err != nil {
		app.sendServerErrorResponse(w, r, errors.Wrap(err, "error submitting incidence report"))
		return
	}

	app.sendAPIResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message: "Your report has been submitted successfully. " +
			"Our investigation partners will look into your report swiftly",
	}, r, nil)
	app.notificationHub.Dispatch(model.NewIncidenceReportNotification(report.UserID))
}

var wsUpgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	CheckOrigin:      func(r *http.Request) bool { return true },
}

// notifications upgrades an http connection to websocket and dispatches
// unread model.Notification messages to the client immediately after a successful connection.
//
// Clients can request for unread connections by sending the text getAllUnread.
//
// Clients can also mark a notification as read by sending the text read->[notificationID],
// e.g., read:qw124fdifhe848skdi3s, which consequentially prompts the server
// to delete such notification from storage.
// The above implies that this server doesn't persist notifications that has been
// read by the client. So clients should take on the responsibility of persisting such.
func (app *app) notifications(w http.ResponseWriter, r *http.Request) {

	// check that userId in url path is valid
	urlPaths := strings.Split(r.URL.Path, "/")
	userId := urlPaths[len(urlPaths)-1]
	if err := app.repo.IsValidUser(userId); err != nil {
		if err == db.ErrUserNotFound {
			app.sendMethodNotAllowedResponse(w, r)
			return
		}
	}

	// upgrade http connection
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Logger.LogError("error upgrading connection to websocket", "dispatch notification", err)
		return
	}

	// send notifications
	logger.Logger.LogServe(200, r)
	app.notificationHub.AddConnection(userId, conn)
	app.notificationHub.DispatchAllUnread(userId)

	for {

		// Read from connection indefinitely to detect closed connection.
		// From gorilla websocket documentation, messageType is either TextMessage or BinaryMessage.
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			logger.Logger.LogError(fmt.Sprintf("removing websocket connection for %s", userId),
				"reading from websocket connection", err)
			app.notificationHub.RemoveConnection(userId)
			break
		}
		if msgType == websocket.TextMessage && string(msg) == "getAllUnread" {
			app.notificationHub.DispatchAllUnread(userId)
			continue
		}

		msgParts := strings.Split(string(msg), ":")
		if msgType == websocket.TextMessage && len(msgParts) > 1 {

			notificationID := msgParts[1]
			app.notificationHub.storage.ReadNotification(notificationID)
		}

	}
}
