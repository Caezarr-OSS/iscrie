package maven2

import (
	"fmt"
	"iscrie/core/importer"
	"iscrie/utils"
)

// Maven2Error representes specific errors for Maven2 repository.
type Maven2Error struct {
	importer.ImportError        // Use generic values
	GroupID              string `json:"group_id,omitempty"`
	ArtifactID           string `json:"artifact_id,omitempty"`
	Version              string `json:"version,omitempty"`
	Classifier           string `json:"classifier,omitempty"`
}

// NewMaven2Error creates Maven2Error instances.
func NewMaven2Error(filePath, groupID, artifactID, version, classifier, errorMessage string) Maven2Error {
	return Maven2Error{
		ImportError: importer.ImportError{
			FilePath:       filePath,
			RepositoryType: "maven2",
			Error:          errorMessage,
		},
		GroupID:    groupID,
		ArtifactID: artifactID,
		Version:    version,
		Classifier: classifier,
	}
}

func FormatMaven2ErrorMessage(e Maven2Error) string {
	return fmt.Sprintf(
		"Maven2 Error - File: %s, GroupID: %s, ArtifactID: %s, Version: %s, Classifier: %s, Error: %s",
		e.FilePath, e.GroupID, e.ArtifactID, e.Version, e.Classifier, e.ImportError.Error,
	)
}

func (e Maven2Error) Error() string {
	formattedMessage := FormatMaven2ErrorMessage(e)
	utils.LogError(formattedMessage)
	return formattedMessage
}
