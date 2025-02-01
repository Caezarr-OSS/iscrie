package maven2

import (
	"fmt"
	"iscrie/utils"
	"path/filepath"
	"regexp"
	"strings"
)

// ParseMavenFileName parses the Maven file name and extracts artifact details.
func ParseMavenFileName(fileName string, versionFromPath string) (artifactID, version, classifier, extension string, err error) {
	extension = filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, extension)

	regex := regexp.MustCompile(`^(.+?)-(\d[\w\.-]*?)(?:-([\w\.-]+))?$`)
	matches := regex.FindStringSubmatch(baseName)

	if matches == nil || len(matches) < 3 {
		return "", "", "", "", utils.LogAndReturnError("failed to parse file name '%s'. Regex did not match", fileName)
	}

	artifactID = matches[1]
	parsedVersion := matches[2]
	classifierCandidate := matches[3]

	if parsedVersion == versionFromPath {
		version = parsedVersion
		classifier = classifierCandidate
	} else {
		version = versionFromPath
		classifier = parsedVersion
	}

	return artifactID, version, classifier, extension, nil
}

// GenerateMavenPath generates the path for the artifact in Maven repository format.
func GenerateMavenPath(groupID, artifactID, version, classifier, extension string) string {
	basePath := strings.ReplaceAll(groupID, ".", "/") + "/" + artifactID + "/" + version

	var formattedPath string
	if classifier != "" {
		formattedPath = fmt.Sprintf("%s/%s-%s-%s%s", basePath, artifactID, version, classifier, extension)
	} else {
		formattedPath = fmt.Sprintf("%s/%s-%s%s", basePath, artifactID, version, extension)
	}
	utils.LogDebug("Generated Maven Path: %s", formattedPath)
	return formattedPath
}
