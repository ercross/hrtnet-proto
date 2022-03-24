package main

import (
	"fmt"
	"github.com/Hrtnet/social-activities/internal/logger"
	"github.com/pkg/errors"
	"net/http"
	"runtime/debug"
	"time"
)

// attachRequestTime set a Time-Received header value for request.
// The value is specified as time.RFC3339
func attachRequestTime(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		t := time.Now().Format(time.RFC3339)
		request.Header.Set("Time-Received", t)
		next.ServeHTTP(writer, request)
	})
}

// recoverer recovers from panic attacks and logs warning
func (app *app) recoverer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil && rvr != http.ErrAbortHandler {
				logger.Logger.LogWarn(string(debug.Stack()), fmt.Sprintf("%v", rvr), errors.New("panic error"))
				app.sendServerErrorResponse(w, r, errors.New("****** panic recovering ******"))
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
