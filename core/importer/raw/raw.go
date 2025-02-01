package raw

import (
	"fmt"
	"iscrie/config"
	"iscrie/core/importer"
	"iscrie/network"
	"iscrie/utils"
	"path/filepath"
)

// RawImporter handles importing RAW files into Nexus.
type RawImporter struct {
	BaseURL      string
	Repository   string
	HTTPClient   *network.HTTPClientAdapter // ✅ Convertit en HTTPClientAdapter
	RootPath     string
	ForceReplace bool
	Config       *config.Config
}

// NewRawImporter creates a new RawImporter instance.
func NewRawImporter(baseURL, repository, rootPath string, httpClient *network.HTTPClient, forceReplace bool) *RawImporter {
	adapter := network.NewHTTPClientAdapter(httpClient, baseURL, repository, forceReplace) // ✅ Adaptation

	return &RawImporter{
		BaseURL:      baseURL,
		Repository:   repository,
		HTTPClient:   adapter, // ✅ Stocke le HTTPClientAdapter
		RootPath:     rootPath,
		ForceReplace: forceReplace,
	}
}

// BuildTargetURL constructs the target URL for RAW files.
func (ri *RawImporter) BuildTargetURL(filePath string) (string, error) {
	if filePath == "" {
		return "", NewRawError(filePath, "", "file path cannot be empty")
	}

	normalizedPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", NewRawError(filePath, "", fmt.Sprintf("failed to normalize path: %v", err))
	}

	relativePath, err := filepath.Rel(ri.RootPath, normalizedPath)
	if err != nil {
		return "", NewRawError(normalizedPath, "", fmt.Sprintf("failed to compute relative path: %v", err))
	}

	return fmt.Sprintf("%srepository/%s/%s", utils.NormalizeBaseURL(ri.BaseURL), ri.Repository, filepath.ToSlash(relativePath)), nil
}

// UploadRawFile uploads a RAW file to Nexus with retry logic.
func (ri *RawImporter) UploadRawFile(filePath string, retryAttempts int, debugLogger, errorLogger func(format string, args ...interface{})) error {
	// Default no-op loggers if nil
	if debugLogger == nil {
		debugLogger = func(format string, args ...interface{}) {}
	}
	if errorLogger == nil {
		errorLogger = func(format string, args ...interface{}) {}
	}

	// Step 1: Build the target URL
	targetURL, err := ri.BuildTargetURL(filePath)
	if err != nil {
		errorLogger("Failed to build target URL for file '%s': %v", filePath, err)
		return fmt.Errorf("failed to build target URL: %w", err)
	}

	// Step 2: Call `UploadFileWithRetry` with both loggers
	return importer.UploadFileWithRetry(ri.HTTPClient, targetURL, filePath, retryAttempts, debugLogger, errorLogger)
}
