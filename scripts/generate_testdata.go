package scripts

import (
	"iscrie/utils"
	"os"
	"path/filepath"
)

// GenerateTestData creates the test data directory structure and files.
func GenerateTestData(root string) error {
	artifacts := []struct {
		path    string
		content string
	}{
		{"com/example/artifact/1.0.0/artifact-1.0.0.jar", "Dummy JAR content"},
		{"com/example/artifact/1.0.0/artifact-1.0.0.pom", "Dummy POM content"},
		{"com/example/artifact/1.0.0/artifact-1.0.0-sources.jar", "Dummy Sources content"},
		{"com/example/artifact/1.0.0/artifact-1.0.0-tests.war", "Dummy WAR content"},
		{"com/example/artifact/1.0.1/artifact-1.0.1-PRE-RC1-SNAPSHOT.jar", "Dummy SNAPSHOT JAR content"},
		{"com/example/artifact/1.0.1/artifact-1.0.1-PRE-RC1.pom", "Dummy POM content"},
		{"com/example/artifact/1.0.1/artifact-1.0.1-TEST.war", "Dummy TEST WAR content"},
		{"com/example/artifact/1.0.1/artifact-1.0.1.ear", "Dummy EAR content"},
		{"com/example/complex-name/2.0.0/complex-name-2.0.0.jar", "Complex JAR content"},
		{"com/example/complex-name/2.0.0/complex-name-2.0.0.pom", "Complex POM content"},
		{"com/example/complex-name/2.0.0/complex-name-2.0.0-SNAPSHOT.zip", "Complex SNAPSHOT ZIP content"},
		{"other/random/files/somefile.txt", "Random file content"},
		{"other/random/files/anotherfile.docx", "Another random file content"},
	}

	for _, artifact := range artifacts {
		filePath := filepath.Join(root, artifact.path)
		dir := filepath.Dir(filePath)

		// Create the directory if it doesn't exist
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return utils.LogAndReturnError("Failed to create directory %s: %v", dir, err)
		}

		// Create the file and write the content
		if err := os.WriteFile(filePath, []byte(artifact.content), os.ModePerm); err != nil {
			return utils.LogAndReturnError("Failed to create file %s: %v", filePath, err)
		}

		utils.LogInfo("Created file: %s", filePath)
	}

	utils.LogInfo("Test data generated successfully in: %s", root)
	return nil
}
