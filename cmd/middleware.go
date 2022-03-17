package main

import (
	"net/http"
)

func requestLogger(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// todo implement requestLogger
		next.ServeHTTP(w, r)
	})
}
