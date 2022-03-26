module github.com/Hrtnet/social-activities

go 1.17

// Below are build constraints syntax to pass configs to Heroku.
// Check https://github.com/heroku/heroku-buildpack-go#go-module-specifics for more info
// +heroku goVersion go1.17

// This specifies the directory where main package lives
// +heroku install ./cmd/...

require (
	github.com/go-chi/chi/v5 v5.0.7
	github.com/go-chi/cors v1.2.0
	github.com/gorilla/websocket v1.5.0
	github.com/jakoubek/onetimecode v0.2.4
	github.com/joho/godotenv v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	go.mongodb.org/mongo-driver v1.8.3
	go.uber.org/zap v1.21.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)

require (
	github.com/BurntSushi/toml v1.0.0 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.0.2 // indirect
	github.com/xdg-go/stringprep v1.0.2 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.0.0-20201216223049-8b5274cf687f // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/text v0.3.5 // indirect
)
