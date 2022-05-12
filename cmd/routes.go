package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"time"
)

func (app *app) routes() http.Handler {
	corsOptions := cors.Options{
		AllowedOrigins:   []string{"http://*", "https://*"}, // Use this to allow specific origin hosts
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}

	mux := chi.NewRouter()
	mux.Use(attachRequestTime)
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(app.recoverer)
	mux.Use(middleware.AllowContentType("application/json", "application/gzip", "multipart/form-data"))
	mux.Use(middleware.SetHeader("Content-Type", "application/json"))
	mux.Use(middleware.Timeout(30 * time.Second))
	mux.Use(cors.Handler(corsOptions))

	mux.Get("/api/health", app.checkStatus)
	mux.Get("/res/images/*", app.serveImages)
	mux.Get("/api/activities-statistics", app.serveAllAirdropSubmission)
	mux.Get("/api/new-user", app.serveStarterPack)
	mux.Get("/api/qr-code", app.serveQrCode)
	mux.Get("/api/task-report", app.serveAirdropSubmission)
	mux.Get("/api/wallet-address", app.serveWalletAddress)
	mux.Get("/api/user/{uid}", app.serveUserInfo)
	mux.Get("/api/announcements", app.serveAnnouncements)
	mux.Get("/api/notifications/{user_id}", app.notifications)

	mux.Post("/api/incidence-report", app.submitIncidenceReport)
	mux.Post("/api/task-report", app.submitAirdropForm)
	mux.Post("/api/validate-qr", app.validateQrCode)
	mux.Post("/api/validate-code", app.validateShortCode)
	mux.Post("/api/validate-rfid", app.validateRFIDText)
	mux.Post("/api/contact-us", app.submitContactUsMessage)
	mux.Post("/api/update-user", app.updateUser)
	mux.Post("/api/reward-alert", app.sendRewardsAlert)
	mux.Post("/api/report-status", app.submitIncidenceReportStatus)
	mux.Post("/api/announcement", app.submitAnnouncement)

	mux.MethodNotAllowed(app.sendMethodNotAllowedResponse)
	mux.NotFound(app.sendNotFoundResponse)
	return mux
}

func (app *app) serveImages(w http.ResponseWriter, r *http.Request) {

	file := r.URL.Path

	fileBytes, err := ioutil.ReadFile("." + file)
	if err != nil {
		app.sendServerErrorResponse(w, r,
			errors.Wrap(err, fmt.Sprintf("failed to read file %s", file)))
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.Write(fileBytes)
	return

}
