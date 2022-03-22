package main

import (
	"fmt"
	"github.com/Hrtnet/social-activities/internal/db"
	"github.com/Hrtnet/social-activities/internal/logger"
	"github.com/Hrtnet/social-activities/internal/model"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/skip2/go-qrcode"
	"net/http"
	"time"
)

// serveAllTaskReports serves activities statistics
// METHOD: GET
// Request must contain admin authorization
func (app *app) serveAllTaskReports(w http.ResponseWriter, r *http.Request) {
	reports, err := app.repo.FetchAllTaskReports()
	if err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}

	app.sendAPIResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message:    "All participants' reports",
	}, reports)
	return
}

// serveTaskReport
// returns the task report submitted by user identified by
// user_id in query parameter
// METHOD: GET
// Query parameter: user_id
func (app *app) serveTaskReport(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("user_id")
	report, err := app.repo.FetchTaskReportByUserID(userId)
	if err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}

	app.sendAPIResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message:    "Welcome to HeartNet",
	}, report)
	return
}

// serveTasks
// Serves tasks to be completed by airdrop participants
// for HeartNet airdrop program
// METHOD: GET
// Content-type: application/json
func (app *app) serveTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := app.repo.FetchAllTasks()
	if err != nil {
		app.sendServerErrorResponse(w, r, errors.New("error fetching tasks"))
		return
	}
	app.sendAPIResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message:    "Welcome to HeartNet",
	}, tasks)
	return
}

// submitTaskReport
// METHOD: POST
// Accept: application/json
// Request Body fields
// 			telegram_username string *required
//			twitter_username string *required
//			tweet_link string *required
// 			youtube_username string *required
// 			wallet_address string (must be present if email_address is absent
//			email_address string (must be present if wallet_address is absent
// 			user_id string *required
func (app *app) submitTaskReport(w http.ResponseWriter, r *http.Request) {

	var report model.TasksReport
	err := app.readJSON(w, r, &report)
	if err != nil {
		app.sendBadRequestResponse(w, r, err)
		return
	}

	if errs := validateTaskReport(&report); errs != nil {
		app.sendFailedValidationResponse(w, r, errs)
		return
	}

	if err = app.repo.CreateTaskReport(&report); err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}

	app.sendAPIResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message:    "Your participation has been recorded successfully",
	}, nil)
	return
}

// validateQrCode
// Method: POST
// Content-Type application/json

// Request Body fields
// data string *required (the text resulting from QR code scan)
func (app *app) validateQrCode(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Data string `json:"data"`
	}
	err := app.readJSON(w, r, &in)
	if err != nil {
		app.sendBadRequestResponse(w, r, err)
		return
	}

	qr := model.DrugFromString(in.Data)
	if qr == nil {
		errs := make(map[string]string, 1)
		errs["error"] = "invalid QR Code"
		app.sendFailedValidationResponse(w, r, errs)
		return
	}

	drug, err := app.repo.ValidateQrText(in.Data)
	if err != nil {
		errs := make(map[string]string, 1)
		errs["error"] = err.Error()
		app.sendFailedValidationResponse(w, r, errs)
		return
	}

	app.sendAPIResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message:    "Valid Drug",
	}, drug)
}

// validateShortCode
// Method: POST
// Accept application/json
// Request Body
// 		data string *required (the short code)
func (app *app) validateShortCode(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Data string `json:"data"`
	}
	err := app.readJSON(w, r, &in)
	if err != nil {
		app.sendBadRequestResponse(w, r, err)
		return
	}

	drug, err := app.repo.ValidateShortCode(in.Data)
	if err != nil {
		errs := make(map[string]string, 1)
		errs["error"] = err.Error()
		app.sendFailedValidationResponse(w, r, errs)
		return
	}

	app.sendAPIResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message:    "Valid Drug",
	}, drug)
}

// validateShortCode
// Method: POST
// Accept application/json
// Request Body Fields
// 		data string *required (the string read from the RFID tag)
func (app *app) validateRFIDText(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Data string `json:"data"`
	}
	err := app.readJSON(w, r, &in)
	if err != nil {
		app.sendBadRequestResponse(w, r, err)
		return
	}

	drug, err := app.repo.ValidateShortCode(in.Data)
	if err != nil {
		errs := make(map[string]string, 1)
		errs["error"] = err.Error()
		app.sendFailedValidationResponse(w, r, errs)
		return
	}

	app.sendAPIResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message:    "Valid Drug",
	}, drug)
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
	return
}

// serveStarterPack serves new user with userId and a wallet address
// METHOD: GET
// Content-Type: application/json
func (app *app) serveStarterPack(w http.ResponseWriter, r *http.Request) {
	userId, err := app.repo.FetchUserID()
	if err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}

	app.sendAPIResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message:    "Welcome to HeartNet",
	}, map[string]interface{}{
		"user_id": userId,
	})
	return
}

func (app *app) checkStatus(w http.ResponseWriter, r *http.Request) {
	app.sendAPIResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message:    "HeartNet prototype is up and running",
	}, nil)
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

	addr, err := app.repo.FetchWalletAddress(userId)
	if err == db.ErrUserNotFound {
		app.sendAPIResponse(&responseWriterArgs{
			writer:     w,
			statusCode: 404,
			status:     false,
			message:    "user id not found",
		}, nil)
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
	}, map[string]string{"address": addr})
}

func (app *app) submitIncidenceReport(w http.ResponseWriter, r *http.Request) {

	// Max memory::32 MB
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		app.sendBadRequestResponse(w, r, err)
		return
	}

	//report, errs := extractIncidenceReport(r)
	//if len (errs) != 0 {
	//	app.sendFailedValidationResponse(w, r, errs)
	//	return
	//}

	if err := app.repo.SubmitIncidenceReport(&model.IncidenceReport{
		ID:                "Temp",
		UserID:            "QWERTY",
		PharmacyName:      "Temp",
		Description:       "Auto generated description",
		PharmacyLocation:  "Auto generated Location",
		EvidenceImagesUrl: nil,
		ReceiptImageUrl:   "",
		Submitted:         time.Now(),
	}); err != nil {
		app.sendServerErrorResponse(w, r, errors.Wrap(err, "error submitting incidence report"))
		return
	}

	app.sendAPIResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message: "Your report has been submitted successfully. " +
			"Our investigation partners will look into your report swiftly",
	}, nil)
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (app *app) dispatchNotifications(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Logger.LogError("error upgrading connection to websocket", "dispatch notification", err)
		return
	}

	err = conn.WriteJSON(struct {
		Message string `json:"message"`
	}{
		Message: "Don't do this",
	})
	if err != nil {
		fmt.Println("error encountered:: ", err)
	}
	//for {
	//	messageType, p, err := conn.ReadMessage()
	//	if err != nil {
	//		log.Println(err)
	//		return
	//	}
	//	if err := conn.WriteMessage(messageType, p); err != nil {
	//		log.Println(err)
	//		return
	//	}
	//}
}
