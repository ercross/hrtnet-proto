package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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

	mux.Get("/api/activities-statistics", app.serveAllTaskReports)
	mux.Get("/api/new-user", app.serveStarterPack)
	mux.Get("/api/qr-code", app.serveQrCode)
	mux.Get("/api/task-report", app.serveTaskReport)
	mux.Get("/api/tasks", app.serveTasks)
	mux.Get("/api/wallet-address", app.serveWalletAddress)
	mux.Post("/api/incidence-report", app.submitIncidenceReport)
	mux.Post("/api/task-report", app.submitTaskReport)
	mux.Post("/api/validate-qr", app.validateQrCode)
	mux.Post("/api/validate-code", app.validateShortCode)
	mux.Post("/api/validate-rfid", app.validateRFIDText)

	mux.Get("/api/notifications/{user_id}", app.notifications)

	mux.MethodNotAllowed(app.sendMethodNotAllowedResponse)
	mux.NotFound(app.sendNotFoundResponse)
	return mux
}
