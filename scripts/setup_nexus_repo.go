package scripts

import (
	"bufio"
	"bytes"
	"encoding/json"
	"iscrie/utils"
	"net/http"
	"os"
	"strings"
)

// RepositoryConfig represents the configuration for a Nexus repository.
type RepositoryConfig struct {
	Name        string `json:"name"`
	Type        string `json:"type"`    // e.g., "maven2" or "raw"
	Format      string `json:"format"`  // e.g., "hosted"
	Version     string `json:"version"` // e.g., "RELEASE" or "MIXED" for maven2
	URL         string `json:"url"`
	User        string `json:"user"`
	Password    string `json:"password"`
	BlobStore   string `json:"blobStore"`   // e.g., "default"
	WritePolicy string `json:"writePolicy"` // e.g., "ALLOW" or "ALLOW_ONCE"
}

// SetupNexusRepo creates a repository in Nexus.
func SetupNexusRepo(config RepositoryConfig) error {
	// Determine the API endpoint based on the repository type and format
	var apiEndpoint string
	switch config.Type {
	case "maven2":
		apiEndpoint = config.URL + "/service/rest/v1/repositories/maven/hosted"
	case "raw":
		apiEndpoint = config.URL + "/service/rest/v1/repositories/raw/hosted"
	default:
		return utils.LogAndReturnError("Unsupported repository type: %s", config.Type)
	}

	// Repository payload
	payload := map[string]interface{}{
		"name":   config.Name,
		"online": true,
		"storage": map[string]interface{}{
			"blobStoreName":               config.BlobStore,
			"strictContentTypeValidation": true,
			"writePolicy":                 config.WritePolicy,
		},
	}

	// Add Maven-specific configuration if needed
	if config.Type == "maven2" {
		payload["maven"] = map[string]interface{}{
			"versionPolicy":      config.Version, // RELEASE or MIXED
			"layoutPolicy":       "STRICT",
			"contentDisposition": "ATTACHMENT",
		}
	}

	// Serialize payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return utils.LogAndReturnError("Failed to serialize payload: %v", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return utils.LogAndReturnError("Failed to create request: %v", err)
	}

	// Add authentication and headers
	req.SetBasicAuth(config.User, config.Password)
	req.Header.Set("Content-Type", "application/json")

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return utils.LogAndReturnError("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusCreated {
		return utils.LogAndReturnError("Failed to create repository '%s': status %d", config.Name, resp.StatusCode)
	}

	utils.LogInfo("Repository '%s' successfully created.", config.Name)
	return nil
}

// ChoiceSetupNexusRepo allows the user to choose and set up a Nexus repository.
func ChoiceSetupNexusRepo() {
	repositories := []RepositoryConfig{
		{
			Name:        "example-raw",
			Type:        "raw",
			Format:      "hosted",
			URL:         "http://localhost:8081",
			User:        "admin",
			Password:    "admin",
			BlobStore:   "default",
			WritePolicy: "ALLOW",
		},
		{
			Name:        "example-maven-releases",
			Type:        "maven2",
			Format:      "hosted",
			Version:     "RELEASE",
			URL:         "http://localhost:8081",
			User:        "admin",
			Password:    "admin",
			BlobStore:   "default",
			WritePolicy: "ALLOW",
		},
		{
			Name:        "example-maven-mixed",
			Type:        "maven2",
			Format:      "hosted",
			Version:     "MIXED",
			URL:         "http://localhost:8081",
			User:        "admin",
			Password:    "admin",
			BlobStore:   "default",
			WritePolicy: "ALLOW",
		},
	}

	utils.LogInfo("Choose the type of repository to create:")
	utils.LogInfo("1. Raw")
	utils.LogInfo("2. Maven2 Hosted (RELEASE)")
	utils.LogInfo("3. Maven2 Hosted (MIXED)")

	reader := bufio.NewReader(os.Stdin)
	utils.LogInfo("Enter your choice (1-3): ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	var selectedRepo RepositoryConfig
	switch choice {
	case "1":
		selectedRepo = repositories[0]
	case "2":
		selectedRepo = repositories[1]
	case "3":
		selectedRepo = repositories[2]
	default:
		utils.LogError("Invalid choice. Exiting.")
		return
	}

	// Attempt to create the selected repository
	err := SetupNexusRepo(selectedRepo)
	if err != nil {
		utils.LogError("Failed to create repository '%s': %v", selectedRepo.Name, err)
	} else {
		utils.LogInfo("Repository '%s' created successfully.", selectedRepo.Name)
	}
}
