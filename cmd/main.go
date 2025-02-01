package main

import (
	"bytes"
	"flag"
	"fmt"
	"iscrie/config"
	"iscrie/core/importer/maven2"
	"iscrie/core/importer/raw"
	"iscrie/network"
	"iscrie/utils" // ✅ Import du nouveau logger
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	// Load configuration
	cfg := initializeConfig()

	// Validate repository type
	if !config.IsValidRepositoryType(cfg.Nexus.RepositoryType) {
		utils.LogError("Invalid repository type: %s. Supported types are: %v", cfg.Nexus.RepositoryType, config.SupportedRepositoryTypes)
		os.Exit(1) // Stop the program
	}

	// Initialize logging system
	err := utils.InitLogger(cfg.General.LogPath, cfg.General.LogLevel)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer utils.CloseLogger() // Ensure log file is closed on exit

	utils.LogInfo("Starting Iscrie...")

	// Init HTTP Client and repository verification
	httpClient := initializeHTTPClient(cfg)
	verifyRepository(cfg, httpClient)

	// Importers initialization
	rawImporter, maven2Importer := initializeImporters(cfg, httpClient)

	processFiles(cfg, rawImporter, maven2Importer)

	utils.LogInfo("Processing completed. Check logs for details.")
}

// initializeConfig loads configuration from TOML file and processes flags.
func initializeConfig() *config.Config {
	configPath := flag.String("config", "iscrie.toml", "Path to the configuration file")
	flag.Parse()

	utils.LogInfo("Loading configuration from: %s", *configPath)
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		utils.LogError("Error loading configuration: %v", err)
	}

	return cfg
}

// initializeHTTPClient configures HTTP client with authentication and proxy.
func initializeHTTPClient(cfg *config.Config) *network.HTTPClient {
	httpClient, err := network.NewHTTPClient(cfg.Auth, cfg.Proxy)
	if err != nil {
		log.Fatalf("Failed to initialize HTTP client: %v", err)
	}
	return httpClient
}

// Verify if repository exists in Nexus
func verifyRepository(cfg *config.Config, httpClient *network.HTTPClient) {

	httpClientAdapter := network.NewHTTPClientAdapter(httpClient, cfg.Nexus.URL, cfg.Nexus.Repository, false)

	// Create Nexus Client
	nexusClient, err := network.NewNexusClient(cfg.Nexus.URL, httpClientAdapter)
	if err != nil {
		utils.LogError("Failed to create Nexus client: %v", err)
	}

	exists, err := nexusClient.RepositoryExists(cfg.Nexus.Repository)
	if err != nil {
		utils.LogError("Failed to check repository existence: %v", err)
	}
	if !exists {
		utils.LogError("Repository '%s' does not exist in Nexus", cfg.Nexus.Repository)
	}
	utils.LogDebug("Verified repository '%s' exists in Nexus.", cfg.Nexus.Repository)
}

// initializeImporters init RAW and Maven2 importers.
func initializeImporters(cfg *config.Config, httpClient *network.HTTPClient) (*raw.RawImporter, *maven2.Maven2Importer) {
	rawImporter := raw.NewRawImporter(cfg.Nexus.URL, cfg.Nexus.Repository, cfg.General.RootPath, httpClient, cfg.Nexus.ForceReplace)
	maven2Importer := maven2.NewMaven2Importer(cfg.Nexus.URL, cfg.Nexus.Repository, cfg.General.RootPath, httpClient, cfg.Nexus.ForceReplace)
	return rawImporter, maven2Importer
}

// processFiles walk through files and processes them according to their type.
func processFiles(cfg *config.Config, rawImporter *raw.RawImporter, maven2Importer *maven2.Maven2Importer) {
	utils.LogDebug("Walking through files in: %s", cfg.General.RootPath)

	start := time.Now()
	totalFiles, successfulUploads, failedUploads := 0, 0, 0
	var uploadErrors []error

	err := filepath.WalkDir(cfg.General.RootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			utils.LogError("Error accessing path: %v", err)
			uploadErrors = append(uploadErrors, fmt.Errorf("error accessing path %s: %w", path, err))
			return nil
		}
		if !d.IsDir() {
			totalFiles++
			utils.LogInfo("Processing file: %s", path)

			var uploadErr error
			if cfg.Nexus.RepositoryType == "maven2" {
				utils.LogInfo("Detected Maven2 file: %s", path)
				uploadErr = maven2Importer.UploadMaven2File(path, cfg.Retry.RetryAttempts, utils.LogDebug, utils.LogError)
			} else if cfg.Nexus.RepositoryType == "raw" {
				utils.LogInfo("Detected RAW file: %s", path)
				uploadErr = rawImporter.UploadRawFile(path, cfg.Retry.RetryAttempts, utils.LogDebug, utils.LogError)
			} else {
				utils.LogInfo("Unsupported repository type: %s", cfg.Nexus.RepositoryType)
			}

			if uploadErr != nil {
				failedUploads++
				uploadErrors = append(uploadErrors, fmt.Errorf("failed to upload file %s: %w", path, uploadErr))
				utils.LogError("Error uploading file: %s, error: %v", path, uploadErr)
			} else {
				successfulUploads++
				utils.LogDebug("Successfully uploaded file: %s", path)
			}
		}
		return nil
	})

	if err != nil {
		utils.LogError("Error during file traversal: %v", err)
	}

	duration := time.Since(start)
	utils.LogInfo("Total files processed: %d", totalFiles)
	utils.LogInfo("Successful uploads: %d", successfulUploads)
	utils.LogInfo("Failed uploads: %d", failedUploads)
	utils.LogInfo("Time taken: %s", duration)

	if len(uploadErrors) > 0 {
		utils.LogInfo("The following errors occurred during upload:")
		var errorBuffer bytes.Buffer
		for _, uploadErr := range uploadErrors {
			utils.LogError("%v", uploadErr)
			errorBuffer.WriteString(fmt.Sprintf("%v\n", uploadErr))
		}
		utils.LogInfo("Upload completed with errors:\n%s", errorBuffer.String())
	}

	utils.LogInfo("All files uploaded successfully.")
}
