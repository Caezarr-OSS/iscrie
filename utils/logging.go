package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Log levels
const (
	DebugLevel = "debug"
	InfoLevel  = "info"
	ErrorLevel = "error"
)

var (
	logFile  *os.File
	logLevel string
	logger   *log.Logger
)

// InitLogger initializes the logging system
func InitLogger(logPath string, level string) error {
	// Ensure the logs directory exists
	if err := os.MkdirAll(logPath, os.ModePerm); err != nil {
		return err
	}

	// Create a log file with timestamp
	logFileName := filepath.Join(logPath, "iscrie_"+time.Now().Format("20060102_150405")+".log")
	var err error
	logFile, err = os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	// Set log level
	logLevel = level

	// Create a multi-writer to log both to file and stdout
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger = log.New(multiWriter, "", log.Ldate|log.Ltime)

	// Log initialization
	logger.Println("[INFO] Logger initialized successfully.")
	return nil
}

// CloseLogger closes the log file
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

// LogDebug logs debug messages
func LogDebug(format string, v ...interface{}) {
	if logLevel == DebugLevel {
		logger.Printf("[DEBUG] "+format, v...)
	}
}

// LogInfo logs informational messages
func LogInfo(format string, v ...interface{}) {
	if logLevel == DebugLevel || logLevel == InfoLevel {
		logger.Printf("[INFO] "+format, v...)
	}
}

// LogError logs error messages
func LogError(format string, v ...interface{}) {
	logger.Printf("[ERROR] "+format, v...)
}

// LogAndReturnError logs an error message and returns an error formatted with fmt.Errorf
func LogAndReturnError(format string, v ...interface{}) error {
	LogError(format, v...)          // Log the error
	return fmt.Errorf(format, v...) // Return the error message
}
