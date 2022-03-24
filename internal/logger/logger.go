package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"net/http"
	"os"

	"time"
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
	l.log.Info("********* APP INFO ********", zap.Field{
		Key:    "Message",
		Type:   zapcore.StringType,
		String: message,
	})
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

	// production log writers
	infoLogsProduction := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./logs/info.log",
		MaxSize:    10,
		MaxAge:     30,
		MaxBackups: 2,
		LocalTime:  true,
		Compress:   true,
	})

	warningLogsProduction := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./logs/warning.log",
		MaxSize:    10,
		MaxAge:     30,
		MaxBackups: 2,
		LocalTime:  true,
		Compress:   true,
	})

	errorLogsProduction := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./logs/error.log",
		MaxSize:    10,
		MaxAge:     30,
		MaxBackups: 2,
		LocalTime:  true,
		Compress:   true,
	})

	fatalLogsProduction := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./logs/fatal.log",
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

	var infoLevel infoEnabler
	var warnLevel warnEnabler
	var errorLevel errorEnabler
	var fatalLevel fatalEnabler

	if isProduction {
		return zapcore.NewTee(
			// these links each logger.LogXXX to their respective log file
			// through the internal calling of zap's methods l.Log.XXX.
			zapcore.NewCore(json, infoLogsProduction, infoLevel),
			zapcore.NewCore(json, warningLogsProduction, warnLevel),
			zapcore.NewCore(json, errorLogsProduction, errorLevel),
			zapcore.NewCore(json, fatalLogsProduction, fatalLevel),
		)
	} else {
		return zapcore.NewTee(
			zapcore.NewCore(console, allDevelopmentLogs, zap.InfoLevel),
		)
	}
}

// By default, using zapcore.InfoLevel as zapcore.LevelEnabler will cause
// the info log file to contain error logs as well because the Enabled method
// for zapcore.InfoLevel returns true for all higher (warn, error, and fatal) levels
type infoEnabler int8
type warnEnabler int8
type errorEnabler int8
type fatalEnabler int8

func (i infoEnabler) Enabled(level zapcore.Level) bool {
	if level == zapcore.InfoLevel {
		return true
	}
	return false
}

func (w warnEnabler) Enabled(level zapcore.Level) bool {
	if level == zapcore.WarnLevel {
		return true
	}
	return false
}

func (e errorEnabler) Enabled(level zapcore.Level) bool {
	if level == zapcore.ErrorLevel {
		return true
	}
	return false
}

func (f fatalEnabler) Enabled(level zapcore.Level) bool {
	if level == zapcore.FatalLevel {
		return true
	}
	return false
}

// LogServe logs http.Request and necessary info about the response for the request
func (l *logger) LogServe(statusCode int, request *http.Request) {
	received := request.Header.Get("Time-Received")
	var duration time.Duration
	if received != "" {
		pt, err := time.Parse(time.RFC3339, received)
		if err != nil {
			l.LogError("error parsing request header value Time-Received",
				"logging server request/response", err)
		}
		duration = time.Since(pt)
	} else {
		duration = time.Since(time.Now())
	}

	l.log.Info("********** SERVER ***********",
		zap.Field{
			Key:    "IP-Addr",
			Type:   zapcore.StringType,
			String: request.RemoteAddr,
		},
		zap.Field{
			Key:    "Method",
			Type:   zapcore.StringType,
			String: request.Method,
		},
		zap.Field{
			Key:    "Path",
			Type:   zapcore.StringType,
			String: request.URL.Path,
		},
		zap.Field{
			Key:    "Status Code",
			Type:   zapcore.StringType,
			String: fmt.Sprintf("%d", statusCode),
		},
		zap.Field{
			Key:    "Status",
			Type:   zapcore.StringType,
			String: http.StatusText(statusCode),
		},
		zap.Field{
			Key:    "Duration",
			Type:   zapcore.StringType,
			String: duration.String(),
		},
	)
}
