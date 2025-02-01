package importer

import (
	"fmt"
	"iscrie/network"
	"iscrie/network/middleware"
	"net/http"
	"time"
)

func UploadFileWithRetry(
	uploader *network.HTTPClientAdapter,
	fullURL, filePath string,
	retryAttempts int,
	debugLogger, errorLogger func(format string, args ...interface{}),
) error {
	return middleware.Retry(retryAttempts, 2*time.Second, func() error {
		// Step 1 : constructs PUT request
		req, file, err := uploader.CreatePutRequest(fullURL, filePath)
		if err != nil {
			errorLogger("Failed to prepare request for file '%s': %v", filePath, err)
			return fmt.Errorf("failed to prepare request for file '%s': %w", filePath, err)
		}
		defer file.Close()

		// Step 2 : executes HTTP request via adapter
		resp, err := uploader.Do(req)
		if err != nil {
			errorLogger("Failed to upload file '%s': %v", filePath, err)
			return fmt.Errorf("failed to upload file '%s': %w", filePath, err)
		}
		defer resp.Body.Close()

		// Step 3 : verify HTTP status
		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			errorLogger("Unexpected response status %d for file '%s'", resp.StatusCode, filePath)
			return fmt.Errorf("unexpected response status %d for file '%s'", resp.StatusCode, filePath)
		}

		debugLogger("Successfully uploaded file: %s", filePath)
		return nil
	})
}
