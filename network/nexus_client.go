package network

import (
	"errors"
	"fmt"
	"iscrie/utils"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// NexusClient represents a client to interact with the Nexus Registry API.
type NexusClient struct {
	BaseURL    string
	HTTPClient *HTTPClientAdapter
}

// NewNexusClient creates a new NexusClient instance with authentication and optional proxy support.
func NewNexusClient(baseURL string, httpClient *HTTPClientAdapter) (*NexusClient, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("baseURL cannot be empty")
	}

	utils.LogDebug("Initializing NexusClient with BaseURL: %s", baseURL)
	return &NexusClient{
		BaseURL:    utils.NormalizeBaseURL(baseURL),
		HTTPClient: httpClient,
	}, nil
}

// RepositoryExists checks if a repository exists in Nexus.
func (c *NexusClient) RepositoryExists(repository string) (bool, error) {
	if repository == "" {
		return false, errors.New("repository name cannot be empty")
	}

	url := fmt.Sprintf("%sservice/rest/v1/repositories/%s", c.BaseURL, repository)
	utils.LogDebug("Checking if repository exists: %s", url)

	// Create a HEAD request to check if the repository exists
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return false, utils.LogAndReturnError("Failed to create HEAD request: %w", err)
	}

	// Execute the request using HTTPClientAdapter
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return false, utils.LogAndReturnError("Failed to execute repository existence check: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	switch resp.StatusCode {
	case http.StatusOK:
		utils.LogInfo("Repository '%s' exists in Nexus.", repository)
		return true, nil
	case http.StatusNotFound:
		utils.LogError("Repository '%s' does not exist in Nexus.", repository)
		return false, nil
	default:
		return false, utils.LogAndReturnError("Unexpected response status %d when checking repository existence", resp.StatusCode)
	}
}

// UploadFile uploads a file to the specified repository in Nexus.
func (c *NexusClient) UploadFile(repository, filePath string, forceReplace bool) error {
	if repository == "" {
		return errors.New("repository name cannot be empty")
	}

	// Construct the upload URL
	relativePath := filepath.ToSlash(filePath)
	url := fmt.Sprintf("%srepository/%s/%s", c.BaseURL, repository, relativePath)
	utils.LogDebug("Uploading file to Nexus: %s -> %s", filePath, url)

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return utils.LogAndReturnError("Failed to open file '%s': %w", filePath, err)
	}
	defer file.Close()

	// Create a PUT request
	req, err := http.NewRequest(http.MethodPut, url, file)
	if err != nil {
		return utils.LogAndReturnError("Failed to create PUT request for file '%s': %w", filePath, err)
	}

	// Add common headers
	AddCommonHeaders(req, forceReplace)

	// Execute the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return utils.LogAndReturnError("Failed to execute upload request for file '%s': %w", filePath, err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return utils.LogAndReturnError("Upload failed with status %d for file '%s'", resp.StatusCode, filePath)
	}

	utils.LogInfo("Successfully uploaded file: %s", filePath)
	return nil
}

// UploadFileWithRetry retries uploading a file to the specified repository in Nexus.
func (c *NexusClient) UploadFileWithRetry(repository, filePath string, forceReplace bool, retryAttempts int, backoff time.Duration) error {
	var lastErr error

	for attempt := 1; attempt <= retryAttempts; attempt++ {
		utils.LogDebug("Attempt %d/%d: Uploading file '%s' to repository '%s'", attempt, retryAttempts, filePath, repository)
		lastErr = c.UploadFile(repository, filePath, forceReplace)
		if lastErr == nil {
			utils.LogInfo("File '%s' successfully uploaded after %d attempt(s)", filePath, attempt)
			return nil // Success
		}

		utils.LogError("Attempt %d/%d failed: %v", attempt, retryAttempts, lastErr)
		time.Sleep(backoff)
		backoff *= 2 // Exponential backoff
	}

	return utils.LogAndReturnError("Upload failed after %d attempts for file '%s'", retryAttempts, filePath)
}
