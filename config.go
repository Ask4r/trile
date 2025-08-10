package main

import (
	"log/slog"
	"os"

	"github.com/pkg/errors"
)

type AppConfig struct {
	BotApiKey string
	LogLevel  slog.Level
	LogFile   string
	TmpDir    string
}

func LoadConfig() (*AppConfig, error) {
	botApiKey := os.Getenv("BOT_API_KEY")
	if botApiKey == "" {
		return nil, errors.New("Cannot retrieve environment variable \"BOT_API_KEY\"\n")
	}
	logLevelString := os.Getenv("LOG_LEVEL")
	if logLevelString == "" {
		return nil, errors.New("Cannot retrieve environment variable \"LOG_LEVEL\"\n")
	}
	var logLevel slog.Level
	switch logLevelString {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		return nil, errors.New("Environment variable \"LOG_LEVEL\" must be \"debug\", \"info\", \"warn\" or \"error\", but \"" + logLevelString + "\" was found")
	}
	logFile := os.Getenv("LOG_FILE")
	if logFile == "" {
		return nil, errors.New("Cannot retrieve environment variable \"LOG_FILE\"\n")
	}
	tmpDir := os.Getenv("TMP_DIR")
	if tmpDir == "" {
		return nil, errors.New("Cannot retrieve environment variable \"TMP_DIR\"\n")
	}

	return &AppConfig{
		BotApiKey: botApiKey,
		LogLevel:  logLevel,
		LogFile:   logFile,
		TmpDir:    tmpDir,
	}, nil
}
