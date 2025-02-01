package scripts

import (
	"fmt"
	"iscrie/config"
	"iscrie/utils"
	"os"
)

func RunIntegrationTest() {
	// Charger la configuration pour récupérer le chemin des logs et le niveau
	cfg, err := config.LoadConfig("./iscrie.toml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialiser le logger
	err = utils.InitLogger(cfg.General.LogPath, cfg.General.LogLevel)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer utils.CloseLogger()

	utils.LogInfo("Select an operation:")
	utils.LogInfo("1. Generate Test Data")
	utils.LogInfo("2. Setup Nexus Repository")
	utils.LogInfo("3. Upload Files to Nexus")
	utils.LogInfo("4. Validate Nexus Upload")
	utils.LogInfo("5. Cleanup Test Data")
	utils.LogInfo("6. Exit")

	var choice int
	utils.LogInfo("Enter your choice: ")
	_, err = fmt.Scanln(&choice)
	if err != nil {
		utils.LogError("Invalid input: %v", err)
		os.Exit(1)
	}

	switch choice {
	case 1:
		utils.LogInfo("Generating test data...")
		if err := GenerateTestData("./test/testdata"); err != nil {
			utils.LogError("Error generating test data: %v", err)
		} else {
			utils.LogInfo("Test data generated successfully.")
		}

	case 2:
		utils.LogInfo("Setting up Nexus repository...")
		ChoiceSetupNexusRepo()

	case 3:
		utils.LogInfo("Uploading files to Nexus...")

		configPath := "./iscrie.toml"
		err := UploadTestData(configPath)
		if err != nil {
			utils.LogError("Error uploading files to Nexus: %v", err)
		} else {
			utils.LogInfo("All files uploaded successfully to Nexus.")
		}

	case 4:
		utils.LogInfo("Validating Nexus upload...")

		baseURL := "http://localhost:8081"
		repoName := "test"
		rootPath := "./test/testdata"

		cfg, err := config.LoadConfig("./iscrie.toml")
		if err != nil {
			utils.LogError("Error loading configuration: %v", err)
			return
		}

		err = ValidateNexusUpload(baseURL, repoName, rootPath, cfg)
		if err != nil {
			utils.LogError("Error validating Nexus upload: %v", err)
		} else {
			utils.LogInfo("Nexus upload validation successful.")
		}

	case 5:
		utils.LogInfo("Cleaning up test data...")
		if err := CleanupTestData("./test/testdata"); err != nil {
			utils.LogError("Error cleaning up test data: %v", err)
		} else {
			utils.LogInfo("Test data cleaned up successfully.")
		}

	case 6:
		utils.LogInfo("Exiting...")
		os.Exit(0)

	default:
		utils.LogError("Invalid choice. Exiting...")
		os.Exit(1)
	}
}
