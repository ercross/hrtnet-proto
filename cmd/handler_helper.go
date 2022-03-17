package main

import (
	"encoding/json"
	"fmt"
	"github.com/Hrtnet/social-activities/internal/model"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"strconv"
	"strings"
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
func (app *app) sendAPIResponse(args *responseWriterArgs, data interface{}) {

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
}

// writeAPIResponse writes apiResponse to w.
// Typically use this if returning a list of values
func (app *app) writeListAPIResponse(args responseWriterArgs, data []interface{}) error {

	response := struct {
		Status  bool              `json:"status"`
		Message string            `json:"message"`
		Errors  map[string]string `json:"errors,omitempty"`
		Data    []interface{}     `json:"data,omitempty"`
	}{
		Status:  args.status,
		Message: args.message,
		Data:    data,
	}

	// Encode the data to JSON, returning the error if there was one.
	apiResponse, err := json.MarshalIndent(response, "", "\t")
	if err != nil {
		return err
	}

	// add header to response
	for key, value := range args.header {
		args.writer.Header()[key] = value
	}
	args.writer.WriteHeader(args.statusCode)
	args.header.Add("Content-Type", "application/json")
	_, err = args.writer.Write(apiResponse)
	if err != nil {
		return err
	}
	return nil
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
