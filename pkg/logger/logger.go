package logger

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"time"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
)

const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
)

func init() {
	infoLogger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	errorLogger = log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)
}

func Info(v ...any) {
	message := fmt.Sprint(v...)
	coloredMessage := fmt.Sprintf("%s[INFO] %s%s", colorGreen, message, colorReset)
	infoLogger.Output(2, coloredMessage)
}

func Error(v ...any) {
	message := fmt.Sprint(v...)
	coloredMessage := fmt.Sprintf("%s[ERROR] %s%s", colorRed, message, colorReset)
	errorLogger.Output(2, coloredMessage)
}

func SetupLogger() (*slog.Logger, *os.File) {
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		os.Mkdir("logs", 0o755)
	}

	logFileName := fmt.Sprintf("logs/log_%s.log", time.Now().Format("20060102_150405"))
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		fmt.Printf("Error creating the logs file: %v\n", err)
		os.Exit(1)
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger := slog.New(slog.NewTextHandler(multiWriter, nil))

	logger.Info("Server starting...", "logFile", logFileName)
	return logger, logFile
}
