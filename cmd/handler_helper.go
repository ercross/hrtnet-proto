package main

import (
	"encoding/json"
	"fmt"
	"github.com/Hrtnet/social-activities/internal/db"
	"github.com/Hrtnet/social-activities/internal/logger"
	"github.com/Hrtnet/social-activities/internal/model"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// extractIDFromQueryParam retrieves "id" URL parameter from the current request.
// Note that the parameter key must be "id", else 0 is returned
func (app *app) extractIDFromQueryParam(r *http.Request) (int64, error) {
	idStr := r.URL.Query().Get("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

type responseWriterArgs struct {
	writer     http.ResponseWriter
	statusCode int

	// status specifies if the request is successful
	status  bool
	message string

	// header specifies additional header fields to add
	// to this response. Note that there is no need setting
	// response type application/json as this has already
	// been done by the middleware added in the routes file
	header http.Header
	errors map[string]string
}

// sendAPIResponse writes response to responseWriterArgs.writer.
func (app *app) sendAPIResponse(args *responseWriterArgs, request *http.Request, data interface{}) {

	response := struct {
		Status  bool              `json:"status"`
		Message string            `json:"message"`
		Data    interface{}       `json:"data,omitempty"`
		Errors  map[string]string `json:"errors,omitempty"`
	}{
		Status:  args.status,
		Message: args.message,
		Data:    data,
		Errors:  args.errors,
	}

	// Encode the data to JSON, returning the error if there was one.
	apiResponse, err := json.MarshalIndent(response, "", "\t")
	if err != nil {
		app.sendServerErrorResponse(args.writer, &http.Request{}, err)
		return
	}

	// write header
	if args.header == nil {
		args.header = http.Header{}
		args.header.Add("Content-Type", "application/json")
	} else {
		args.header.Add("Content-Type", "application/json")
	}
	for key, value := range args.header {
		args.writer.Header()[key] = value
	}
	args.writer.WriteHeader(args.statusCode)

	_, err = args.writer.Write(apiResponse)
	if err != nil {
		app.sendServerErrorResponse(args.writer, &http.Request{}, err)
		return
	}
	logger.Logger.LogServe(args.statusCode, request)
}

// readJSON reads request body r and decodes the result into dst.
// Note that dst must be a pointer to a model e.g., &user
func (app *app) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// limit request body size to 1MB.
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(dst)
	if err != nil {
		return formatReadError(err)
	}

	// Check that request body only contains a single JSON value this.
	// Getting anything other than an io.EOF error indicates there's
	// additional data in the request body
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return fmt.Errorf("request body must only contain a single JSON value")
	}

	return nil
}

func formatReadError(err error) error {
	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError
	var invalidUnmarshalError *json.InvalidUnmarshalError

	switch {

	// if err contains type json.SyntaxError
	case errors.As(err, &syntaxError):
		return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

	// if err is of type json.ErrUnexpectedEOF,
	// usually encountered in cases of json syntax error
	case errors.Is(err, io.ErrUnexpectedEOF):
		return errors.New("body contains badly-formed JSON")

	// If Decode encounters a wrong type for the target destination
	case errors.As(err, &unmarshalTypeError):
		if unmarshalTypeError.Field != "" {
			return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
		}
		return fmt.Errorf("body contains badly-formed JSON (at character %d)", unmarshalTypeError.Offset)

	// If the JSON contains a field which cannot be mapped to target destination,
	// then Decode() will now return an error message in the format "json: unknown
	// field "<name>"". We check for this, extract the field name from the error,
	// and interpolate it into our custom error message. Note that there's an open
	// issue at https://github.com/golang/go/issues/29035 regarding turning this
	// into a distinct error type in the future.
	case strings.HasPrefix(err.Error(), "json: unknown field"):
		fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
		return fmt.Errorf("body contains unknown key %s", fieldName)

	// If the request body exceeds 1MB in size the decode will now fail with the
	// error "http: request body too large".
	case err.Error() == "http: request body too large":
		return errors.New("body must not be larger than allowed bytes")

	// if a non-nil pointer is passed to json.Decode().
	case errors.As(err, &invalidUnmarshalError):
		panic(err)

	// For anything else, return the error message as-is.
	default:
		return err
	}
}

// validateAirdropSubmission checks that the POSTed task report
// contains the required field.
// If one or more fields are missing, it reports an error
func validateAirdropSubmission(report *model.AirdropSubmission) (errs map[string]string) {
	return nil
}

func extractAnnouncement(r *http.Request, saveDir, apiUrl string) (*model.Announcement, db.ErrorType, error) {
	announcement := new(model.Announcement)

	// extract other form values
	announcement.CreatedOn = time.Now()
	announcement.Body = r.PostFormValue("text")
	announcement.Title = r.PostFormValue("title")
	validTill, err := time.Parse("01-02-2006", r.PostFormValue("valid_till"))
	if err != nil {
		return nil, db.ValidationError, errors.Wrap(err, "invalid validity date")
	}
	announcement.ValidTill = validTill
	announcement.Url = r.PostFormValue("url")

	// save image if found
	file, header, err := r.FormFile("image")

	if err != nil {

		if err == http.ErrMissingFile {
			return announcement, db.None, nil
		}
		return nil, db.ValidationError, errors.Wrap(err, "invalid announcement image")
	}
	defer file.Close()

	parts := strings.Split(header.Filename, ".")
	fileExtension := parts[len(parts)-1]
	saveAs := fmt.Sprintf("%d_%s.%s", announcement.CreatedOn.Unix(),
		strings.ReplaceAll(announcement.Title, " ", "_"), fileExtension)

	savedFileUrl, err := saveFile(file, saveAs, saveDir)
	if err != nil {
		if err == errFileTooLarge {
			return nil, db.ValidationError, errors.Wrap(err, "failed to save receipt image")
		}
		return nil, db.InternalError, err
	}
	announcement.ImageUrl = fmt.Sprintf("%s%s", apiUrl, savedFileUrl)

	return announcement, db.None, nil
}

// extractIncidenceReport extracts incidence report data from the request.
// Request content-type must be multipart/form-data.
// Receipt file name is saved as userId_unixTime.
// Evidence images are saved in a userId_unixTime directory
func (app *app) extractIncidenceReport(r *http.Request, cfg *config) (*model.IncidenceReport, db.ErrorType, error) {
	report := new(model.IncidenceReport)

	// check that user is valid
	report.UserID = r.PostFormValue("user_id")
	err := app.repo.IsValidUser(report.UserID)
	if err != nil {
		if err == db.ErrUserNotFound {
			return nil, db.ValidationError, err
		}
		return nil, db.InternalError, err
	}

	// extract other form values
	report.SubmittedOn = time.Now()
	report.Description = r.PostFormValue("description")
	report.PharmacyLocation = r.PostFormValue("pharmacy_location")
	report.PharmacyName = r.PostFormValue("pharmacy_name")

	if report.Description == "" || report.PharmacyName == "" || report.PharmacyLocation == "" {
		return nil, db.ValidationError, errors.New("one of description, pharmacy location, or pharmacy name is missing")
	}

	// save receipt
	file, header, err := r.FormFile("receipt")
	if err != nil {
		return nil, db.ValidationError, errors.Wrap(err, "invalid receipt image")
	}
	defer file.Close()
	filenameParts := strings.Split(header.Filename, ".")
	fileExtension := filenameParts[len(filenameParts)-1]
	saveAs := fmt.Sprintf("%s_%d.%s", report.UserID, time.Now().Unix(), fileExtension)
	savedFileUrl, err := saveFile(file, saveAs, cfg.incidenceReportReceiptImagePath)
	if err != nil {
		if err == errFileTooLarge {
			return nil, db.ValidationError, errors.Wrap(err, "failed to save receipt image")
		}
		return nil, db.InternalError, err
	}
	report.ReceiptImageUrl = fmt.Sprintf("%s%s", app.config.apiUrl, savedFileUrl)

	// save evidence images
	zippedFiles, header, err := r.FormFile("evidence_images")
	if err != nil {
		os.Remove(report.ReceiptImageUrl)
		return nil, db.ValidationError, errors.Wrap(err, "invalid evidence images")
	}

	saveDir := fmt.Sprintf("%s/%s_%d", cfg.incidenceReportDrugImagePath, report.UserID, time.Now().Unix())
	savePaths, errType, err := unzipAndSave(zippedFiles, header, saveDir, app.config.apiUrl)
	if err != nil {
		return nil, errType, err
	}
	// todo obtain correct save path
	report.EvidenceImagesUrl = *savePaths
	return report, db.None, nil
}

func (app *app) processValidation(w http.ResponseWriter, r *http.Request, drug *model.Drug, userId string, err error) {
	if err == db.ErrDrugNotFound {
		app.sendDrugNotFoundResponse(w, r)
		app.notificationHub.Dispatch(model.NewValidationNotification(userId, "Drug not found"))
		return
	}

	if err != nil {
		errs := make(map[string]string)
		errs["error"] = err.Error()
		app.sendFailedValidationResponse(w, r, errs)
		return
	}
	app.notificationHub.Dispatch(model.NewValidationNotification(userId, "Drug is authentic"))
	app.sendDrugFoundResponse(w, r, drug)
}

// sendDrugNotFoundResponse sends appropriate response if drug is not found in repo.
func (app *app) sendDrugNotFoundResponse(w http.ResponseWriter, r *http.Request) {
	app.sendAPIResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message:    "Not Found",
	}, r, map[string]string{
		"report_type": "Unsafe",
	})
}

// sendDrugFoundResponse sends safe or expiry product response,
// depending on if product expires in the next 7 days
func (app *app) sendDrugFoundResponse(w http.ResponseWriter, r *http.Request, drug *model.Drug) {

	// check that drug is not expiring in the next 7 days
	if time.Now().Add(time.Hour * 168).After(drug.Expiry) {
		app.sendAPIResponse(&responseWriterArgs{
			writer:     w,
			statusCode: 200,
			status:     true,
			message:    "Expired Drug",
		}, r, map[string]interface{}{
			"report_type": "expired",
			"drug":        drug,
		})
		return
	}

	app.sendAPIResponse(&responseWriterArgs{
		writer:     w,
		statusCode: 200,
		status:     true,
		message:    "Valid Drug",
	}, r, map[string]interface{}{
		"report_type": "safe",
		"drug":        drug,
	})
}

// dispatchNotification sends notification to the websocket connection, conn
func dispatchNotification(notification model.Notification, conn *websocket.Conn) interface{} {
	conn.WriteJSON(notification)
	return nil
}
