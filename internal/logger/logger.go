package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

// Logger must be initialized with NewLogger before use.
// Use intent: Logger should be used to log only events
// that cannot be captured by the logger middleware
var Logger *logger

type logger struct {
	isProduction bool
	log          *zap.Logger
}

func NewLogger(isProduction bool) *logger {
	return &logger{
		isProduction: isProduction,
		log:          zap.New(encoders(isProduction)),
	}
}

// LogInfo logs messages that state what happened
// in the normal processes of the app
// @action specifies what action is being performed when this logger was invoked
func (l *logger) LogInfo(message string) {
	l.log.Info(message)
}

// LogWarn logs  level is used when you have detected an unexpected application problem
// @action specifies what action is being performed when this logger was invoked
func (l *logger) LogWarn(message, action string, err error) {
	l.log.Warn(message, zap.Field{
		Key:    "action",
		Type:   zapcore.StringType,
		String: action,
	}, zap.Field{
		Key:       "error",
		Type:      zapcore.ErrorType,
		Interface: err,
	})
}

func (l *logger) FlushBuffer() {
	l.log.Sync()
}

// LogError logs severe issues stopping functions within the application from operating efficiently,
// like an inability to access a service.
// @action specifies what action is being performed when this logger was invoked
func (l *logger) LogError(message, action string, err error) {
	l.log.Error(message, zap.Field{
		Key:    "action",
		Type:   zapcore.StringType,
		String: action,
	}, zap.Field{
		Key:       "error",
		Type:      zapcore.ErrorType,
		Interface: err,
	})
}

// LogFatal logs catastrophic application’s situation, such as an important function not working
// LogFatal panics after logging the log message
func (l *logger) LogFatal(message, action string, err error) {
	l.log.Fatal(message, zap.Field{
		Key:    "action",
		Type:   zapcore.StringType,
		String: action,
	}, zap.Field{
		Key:       "error",
		Type:      zapcore.ErrorType,
		Interface: err,
	})
	panic(err)
}

func encoders(isProduction bool) zapcore.Core {

	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "timestamp",
		NameKey:        "name",
		CallerKey:      "invoked_by",
		FunctionKey:    "function",
		StacktraceKey:  "stack_trace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 time: 15:04:05.000Z0700"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}

	// infoLog also log for higher log levels:: warn, error, and fatal
	infoLogsProduction := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./logs/info.log",
		MaxSize:    10,
		MaxAge:     30,
		MaxBackups: 2,
		LocalTime:  true,
		Compress:   true,
	})

	allDevelopmentLogs := zapcore.Lock(os.Stdout)

	// encoders
	json := zapcore.NewJSONEncoder(encoderConfig)
	console := zapcore.NewConsoleEncoder(encoderConfig)

	if isProduction {
		return zapcore.NewCore(json, infoLogsProduction, zap.InfoLevel)

	} else {
		return zapcore.NewCore(console, allDevelopmentLogs, zap.InfoLevel)

	}
}
