package main

import (
	"fmt"
	"github.com/Hrtnet/social-activities/internal/logger"
	"net/http"
)

// sendErrorResponse() method is a generic helper for sending JSON-formatted error
// messages to the client with a given status code.
func (app *app) sendErrorResponse(w http.ResponseWriter, r *http.Request, status int, message string, errors map[string]string) {

	// Write the response using the writeAPIResponse() helper. If this happens to return an
	// error then log it, and fall back to sending the client an empty response with a
	// 500 Internal Server Error status code.
	app.sendAPIResponse(&responseWriterArgs{
		writer:     w,
		status:     false,
		statusCode: status,
		message:    message,
		errors:     errors,
	}, r, nil)
}

// The serverErrorResponse() method will be used when the app encounters an
// unexpected problem at runtime. It logs the detailed error message, then uses the
// sendErrorResponse() helper to send a 500 Internal Server Error status code and JSON
// response (containing a generic error message) to the client.
func (app *app) sendServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	logger.Logger.LogError("internal error encountered", "", err)
	message := "the server encountered an error and could not process your request"
	app.sendErrorResponse(w, r, http.StatusInternalServerError, message, nil)
}

// The sendNotFoundResponse() method will be used to send a 404 Not Found status code
// and JSON response to the client.
func (app *app) sendNotFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.sendErrorResponse(w, r, http.StatusNotFound, message, nil)
}

// sendMethodNotAllowedResponse sends a 405 Method Not Allowed
// status code and JSON response to the client.
func (app *app) sendMethodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	app.sendErrorResponse(w, r, http.StatusMethodNotAllowed, message, nil)
}

// sendBadRequestResponse sends a 400 Bad Request status code
// and JSON response to the client.
func (app *app) sendBadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.sendErrorResponse(w, r, http.StatusBadRequest, err.Error(), nil)
}

// sendFailedValidationResponse sends a 422 Unprocessable Entity
// status code and JSON response to the client.
func (app *app) sendFailedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.sendErrorResponse(w, r, http.StatusUnprocessableEntity, "failed validation", errors)
}

// sendEditConflictResponse sends a 409 Conflict status code
// and JSON response to the client.
func (app *app) sendEditConflictResponse(w http.ResponseWriter, r *http.Request, message string) {
	app.sendErrorResponse(w, r, http.StatusConflict, message, nil)
}
