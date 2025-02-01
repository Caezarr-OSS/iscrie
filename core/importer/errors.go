package importer

import (
	"encoding/json"
	"fmt"
	"iscrie/utils"
	"os"
	"path/filepath"
	"sync"
)

// ImportError represents a generic import error.
type ImportError struct {
	FilePath       string `json:"file_path"`
	RepositoryType string `json:"repository_type"`
	Error          string `json:"error"`
}

// ErrorLogger manages import error logs.
type ErrorLogger struct {
	FilePath string
	Mutex    sync.Mutex
}

// NewErrorLogger creates a new ErrorLogger.
func NewErrorLogger(filePath string) *ErrorLogger {
	return &ErrorLogger{
		FilePath: filePath,
	}
}

// LogError adds a new ImportError to the log file.
func (el *ErrorLogger) LogError(importError ImportError) error {
	el.Mutex.Lock()
	defer el.Mutex.Unlock()

	dir := filepath.Dir(el.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		utils.LogError("Failed to create directory for log file: %v", err)
		return fmt.Errorf("failed to create directory for log file: %w", err)
	}

	file, err := os.OpenFile(el.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		utils.LogError("Failed to open error log file: %v", err)
		return fmt.Errorf("failed to open error log file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(importError); err != nil {
		utils.LogError("Failed to write error to log file: %v", err)
		return fmt.Errorf("failed to write error to log file: %w", err)
	}

	return nil
}

// ReadErrors reads all ImportErrors from a log file.
func (el *ErrorLogger) ReadErrors() ([]ImportError, error) {
	el.Mutex.Lock()
	defer el.Mutex.Unlock()

	file, err := os.Open(el.FilePath)
	if err != nil {
		utils.LogError("Failed to open error log file: %v", err)
		return nil, fmt.Errorf("failed to open error log file: %w", err)
	}
	defer file.Close()

	var errors []ImportError
	decoder := json.NewDecoder(file)
	for {
		var importError ImportError
		if err := decoder.Decode(&importError); err != nil {
			if err.Error() == "EOF" {
				break
			}
			utils.LogError("Failed to decode error log: %v", err)
			return nil, fmt.Errorf("failed to decode error log: %w", err)
		}
		errors = append(errors, importError)
	}

	return errors, nil
}
