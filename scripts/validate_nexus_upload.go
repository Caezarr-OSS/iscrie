package scripts

import (
	"iscrie/config"
	"iscrie/core/importer/maven2"
	"iscrie/network"
	"iscrie/utils"
	"net/http"
	"os"
	"path/filepath"
)

// ValidateNexusUpload vérifie que tous les fichiers du répertoire local existent dans le dépôt Nexus.
func ValidateNexusUpload(baseURL, repoName, rootPath string, cfg *config.Config) error {
	utils.LogInfo("Starting validation of files in %s against Nexus repository %s...", rootPath, repoName)

	// Créer une instance du HTTPClient
	httpClient, err := network.NewHTTPClient(cfg.Auth, cfg.Proxy)
	if err != nil {
		return utils.LogAndReturnError("Failed to initialize HTTP client: %w", err)
	}

	// Créer un adaptateur HTTPClientAdapter pour gérer les requêtes
	httpClientAdapter := network.NewHTTPClientAdapter(httpClient, baseURL, repoName, cfg.Nexus.ForceReplace)

	// Créer une instance de Maven2Importer
	mavenImporter := maven2.NewMaven2Importer(baseURL, repoName, rootPath, httpClient, cfg.Nexus.ForceReplace)

	// Résultats de validation
	var missingFiles []string
	totalFiles := 0
	validatedFiles := 0

	// Parcourir les fichiers locaux
	err = filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			utils.LogError("Error accessing path %s: %v", path, err)
			return nil // Continuer avec les autres fichiers
		}
		if d.IsDir() {
			return nil // Ignorer les répertoires
		}

		// Incrémenter le compteur total des fichiers
		totalFiles++

		// Construire l'URL complète en utilisant Maven2Importer
		fullURL, err := mavenImporter.BuildFullTargetURL(path)
		if err != nil {
			utils.LogError("Error generating URL for %s: %v", path, err)
			missingFiles = append(missingFiles, path)
			return nil
		}
		utils.LogDebug("Checking URL: %s", fullURL)

		// Vérifier l'existence du fichier dans Nexus
		req, err := http.NewRequest(http.MethodHead, fullURL, nil)
		if err != nil {
			utils.LogError("Error creating HTTP request for %s: %v", fullURL, err)
			missingFiles = append(missingFiles, path)
			return nil
		}

		resp, err := httpClientAdapter.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			statusCode := 0
			if resp != nil {
				statusCode = resp.StatusCode
			}
			utils.LogError("Missing file in Nexus: %s (HTTP %d)", path, statusCode)
			missingFiles = append(missingFiles, path)
			return nil
		}
		defer resp.Body.Close()

		// Si le fichier est validé, incrémenter le compteur
		validatedFiles++
		utils.LogDebug("Validated file: %s", path)
		return nil
	})

	// Vérification des erreurs lors de la traversée des fichiers
	if err != nil {
		utils.LogError("Error traversing files: %v", err)
		return err
	}

	// Résumé de validation
	utils.LogInfo("\nValidation summary:")
	utils.LogInfo("Total files processed: %d", totalFiles)
	utils.LogInfo("Total files validated: %d", validatedFiles)

	if len(missingFiles) > 0 {
		utils.LogError("Total missing files: %d", len(missingFiles))
		utils.LogError("\nList of missing files:")
		for _, file := range missingFiles {
			utils.LogError("- %s", file)
		}
		return utils.LogAndReturnError("Validation completed with missing files")
	}

	utils.LogInfo("All files validated successfully in Nexus.")
	return nil
}
