package main

import (
	"encoding/json"
	"fmt"
	"github.com/Hrtnet/social-activities/internal/db"
	"github.com/Hrtnet/social-activities/internal/logger"
	"github.com/Hrtnet/social-activities/internal/model"
	"github.com/pkg/errors"
	"io"
	"net/http"
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

// validateTaskReport checks that the POSTed task report
// contains the required field.
// If one or more fields are missing, it reports an error
func validateTaskReport(report *model.TasksReport) (errs map[string]string) {
	return nil
}

// extractIncidenceReport extracts incidence report data from the request.
// Request content-type must be multipart/form-data.
// Any error returned is a client error
func (app *app) extractIncidenceReport(r *http.Request, cfg config) (*model.IncidenceReport, errorType, map[string]string) {
	var report model.IncidenceReport
	errors := make(map[string]string)

	// check that user is valid
	report.UserID = r.PostFormValue("user_id")
	err := app.repo.IsValidUser(report.UserID)
	if err != nil {
		errors["error"] = err.Error()
		return nil, errBadRequest, errors
	}

	// extract other form values
	// todo use package go-playground/mold
	report.Submitted = time.Now()
	report.Description = r.PostFormValue("description")
	report.PharmacyLocation = r.PostFormValue("pharmacy_location")
	report.PharmacyName = r.PostFormValue("pharmacy_name")

	// todo validate with go-playground/validator
	// todo check that the expected files are not absent

	// save receipt
	file, header, err := r.FormFile("receipt")
	if err != nil {
		errors["receipt"] = "invalid file type"
		return nil, errBadRequest, errors
	}
	defer file.Close()
	filenameParts := strings.Split(header.Filename, ".")
	fileExtension := filenameParts[len(filenameParts)-1]
	saveAs := fmt.Sprintf("%s_receipt.%s", report.UserID, fileExtension)
	savedFileUrl, err := saveFile(file, saveAs, cfg.incidenceReportReceiptImagePath)
	if err != nil {
		errors["error"] = err.Error()
		return nil, errInternal, errors
	}
	report.ReceiptImageUrl = savedFileUrl

	zippedFiles, header, err := r.FormFile("evidence_images")
	if err != nil {
		errors["receipt"] = "invalid file type"
		return nil, errBadRequest, errors
	}

	errType, err := unzipAndSave(zippedFiles, header, cfg.incidenceReportDrugImagePath)
	if err != nil {
		errors["error"] = err.Error()
		return nil, errType, errors
	}
	// todo obtain correct save path
	report.EvidenceImagesUrl = cfg.incidenceReportDrugImagePath
	return &report, nil, nil
}

func (app *app) processValidation(w http.ResponseWriter, r *http.Request, drug *model.Drug, err error) {
	if err == db.ErrDrugNotFound {
		app.sendDrugNotFoundResponse(w, r)
		return
	}

	if err != nil {
		errs := make(map[string]string)
		errs["error"] = err.Error()
		app.sendFailedValidationResponse(w, r, errs)
		return
	}

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
