package main

import (
	"flag"
	"fmt"
	"github.com/Hrtnet/social-activities/internal/db"
	"github.com/Hrtnet/social-activities/internal/logger"
	"github.com/Hrtnet/social-activities/internal/model"
	"github.com/joho/godotenv"
	"os"
	"time"

	"net/http"
)

type config struct {
	port                            int
	environment                     model.Environment
	incidenceReportDrugImagePath    string
	incidenceReportReceiptImagePath string
	announcementImagePath           string
	apiUrl                          string
	dsn                             string
}

type app struct {
	config          *config
	repo            Repository
	notificationHub *NotificationHub
}

// createDirs creates necessary file server directories with
// the appropriate file permissions
func createDirs(dirs ...string) {
	for _, dir := range dirs {
		os.MkdirAll(dir, 0700)
	}
}

func main() {

	cfg := initConfig()
	logger.Logger = logger.NewLogger(cfg.environment == model.Production)
	createDirs(
		cfg.announcementImagePath,
		cfg.incidenceReportReceiptImagePath,
		cfg.incidenceReportDrugImagePath)
	model.InitializeFirebaseAdminSDK()
	mongo, err := db.ConnectMongo(cfg.dsn)
	if err != nil {
		logger.Logger.LogFatal("error connecting to database", "", err)
	}
	app := &app{
		config: &cfg,
		repo:   mongo,
	}
	app.notificationHub = NewNotificationHub(mongo)
	defer app.repo.Disconnect()

	app.serve()
}

func (app *app) serve() error {
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.config.port),
		Handler:           app.routes(),
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      20 * time.Second,
		MaxHeaderBytes:    2048,
	}

	logger.Logger.LogInfo(
		fmt.Sprintf("starting server in %s mode on port %d", app.config.environment.String(), app.config.port))
	return server.ListenAndServe()
}

func initConfig() config {
	var config config

	flag.IntVar(&config.port, "port", 4042, "port the server listens on")
	flag.Var(&config.environment, "environment", "application environment, enum: development, production")
	flag.StringVar(&config.apiUrl, "apiUrl", "localhost", "api endpoint")
	flag.Parse()

	if config.environment == model.Development {
		if err := godotenv.Load(); err != nil {
			logger.Logger.LogFatal("failed to load env file", "initializing app config", err)
		}
	}
	config.dsn = os.Getenv("DSN")
	config.incidenceReportDrugImagePath = "./res/images/incidence-reports/drugs"
	config.incidenceReportReceiptImagePath = "./res/images/incidence-reports/receipts"
	config.announcementImagePath = "./res/images/announcements"

	return config
}
