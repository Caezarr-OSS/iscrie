package scripts

import (
	"bytes"
	"iscrie/config"
	"iscrie/core/importer/maven2"
	"iscrie/core/importer/raw"
	"iscrie/network"
	"iscrie/utils"
	"os"
	"path/filepath"
)

// UploadTestData uploads the test data from the specified root directory to the Nexus repositories.
func UploadTestData(configPath string) error {
	// Load configuration from TOML file.
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return utils.LogAndReturnError("Failed to load configuration: %w", err)
	}

	// Verify if root path exists
	rootPath := cfg.General.RootPath
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		return utils.LogAndReturnError("Root path does not exist: %s", rootPath)
	}

	utils.LogInfo("Starting upload of test data from %s...", rootPath)

	// Initialize HTTP client
	httpClient, err := network.NewHTTPClient(cfg.Auth, cfg.Proxy)
	if err != nil {
		return utils.LogAndReturnError("Failed to initialize HTTP client: %w", err)
	}

	// Initialize importers
	rawImporter := raw.NewRawImporter(cfg.Nexus.URL, cfg.Nexus.Repository, rootPath, httpClient, cfg.Nexus.ForceReplace)
	maven2Importer := maven2.NewMaven2Importer(cfg.Nexus.URL, cfg.Nexus.Repository, rootPath, httpClient, cfg.Nexus.ForceReplace)

	// Error list
	var uploadErrors []error

	// Walkthrough root folder
	walkErr := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			uploadErrors = append(uploadErrors, utils.LogAndReturnError("Error accessing path %s: %w", path, err))
			return nil // Continue processing other files
		}

		// Ignore folders
		if d.IsDir() {
			return nil
		}

		// Select importer according to file extension
		var uploadErr error
		if filepath.Ext(path) == ".pom" || filepath.Ext(path) == ".jar" || filepath.Ext(path) == ".war" || filepath.Ext(path) == ".ear" {
			utils.LogDebug("Attempting to upload Maven2 artifact: %s", path)
			uploadErr = maven2Importer.UploadMaven2File(
				path,
				cfg.Retry.RetryAttempts,
				utils.LogDebug,
				utils.LogError,
			)
		} else {
			utils.LogDebug("Attempting to upload RAW file: %s", path)
			uploadErr = rawImporter.UploadRawFile(
				path,
				cfg.Retry.RetryAttempts,
				utils.LogDebug,
				utils.LogError,
			)
		}

		// Handle upload errors
		if uploadErr != nil {
			uploadErrors = append(uploadErrors, utils.LogAndReturnError("Failed to upload file %s: %w", path, uploadErr))
		} else {
			utils.LogInfo("Successfully uploaded: %s", path)
		}

		return nil
	})

	// Check for errors during file walk
	if walkErr != nil {
		return utils.LogAndReturnError("Error during file walk: %w", walkErr)
	}

	// If errors occurred, log them
	if len(uploadErrors) > 0 {
		var errorBuffer bytes.Buffer
		utils.LogError("The following errors occurred during upload:")
		for _, uploadErr := range uploadErrors {
			utils.LogError("%v", uploadErr)
			errorBuffer.WriteString(uploadErr.Error() + "\n")
		}
		return utils.LogAndReturnError("Upload completed with errors:\n%s", errorBuffer.String())
	}

	utils.LogInfo("All test data uploaded successfully.")
	return nil
}
