package app

import (
	"net/url"
	"os"

	"github.com/apsdehal/go-logger"
)

type Logger struct {
	*logger.Logger
}

var log *Logger
var baseLogger *logger.Logger

func (l *Logger) Info(message string) {
	baseLogger.Info(message)
}

func (l *Logger) Debug(message string) {
	if flags.debug {
		baseLogger.Debug(message)
	}
}

func (l *Logger) Verbose(message string) {
	if flags.verbose {
		baseLogger.Info(message)
	}
}

func (l *Logger) Error(message string) {
	baseLogger.Error(message)
}

func (l *Logger) Warning(message string) {
	baseLogger.Warning(message)
}

func (l *Logger) Notice(message string) {
	baseLogger.Notice(message)
}

func (l *Logger) Fatal(message string) {
	if _, err := url.ParseRequestURI(settings.SlackWebhook); err == nil {
		notifySlack(message, settings.SlackWebhook, true, flags.apply)
	}
	baseLogger.Fatal(message)
}

func initLogs(verbose bool, noColors bool) {
	logger.SetDefaultFormat("%{time:2006-01-02 15:04:05} %{level}: %{message}")
	logLevel := logger.InfoLevel
	if verbose {
		logLevel = logger.DebugLevel
	}
	colors := 1
	if noColors {
		colors = 0
	}
	baseLogger, _ = logger.New("logger", colors, os.Stdout, logLevel)
}
