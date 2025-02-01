package raw

import (
	"fmt"
	"iscrie/utils"
)

// RawError represents a specific RAW repository error.
type RawError struct {
	FilePath     string `json:"file_path"`
	RelativePath string `json:"relative_path"`
	BaseError    string `json:"error"`
}

// NewRawError creates a new RawError instance.
func NewRawError(filePath, relativePath, errorMessage string) RawError {
	formattedMessage := FormatRawErrorMessage(filePath, relativePath, errorMessage)
	utils.LogError(formattedMessage)

	return RawError{
		FilePath:     filePath,
		RelativePath: relativePath,
		BaseError:    errorMessage,
	}
}

// FormatRawErrorMessage formats a raw repository error message.
func FormatRawErrorMessage(filePath, relativePath, errorMessage string) string {
	return fmt.Sprintf("Raw Error - File: %s, Relative Path: %s, Error: %s",
		filePath, relativePath, errorMessage)
}

// Error formats a message error specific for Raw repository.
func (e RawError) Error() string {
	return FormatRawErrorMessage(e.FilePath, e.RelativePath, e.BaseError)
}
