package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// ConvertPathsToAbsolute converts a list of paths to absolute paths if they are not already absolute.
func ConvertPathsToAbsolute(paths ...string) ([]string, error) {
	absolutePaths := make([]string, len(paths))
	for i, path := range paths {
		if filepath.IsAbs(path) {
			absolutePaths[i] = NormalizePath(path)
		} else {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return nil, err
			}
			absolutePaths[i] = NormalizePath(absPath)
		}
	}
	return absolutePaths, nil
}

// IsSupportedExtension checks if the given file extension is in the list of supported extensions.
func IsSupportedExtension(ext string, supportedExtensions []string) bool {
	for _, supported := range supportedExtensions {
		if ext == supported {
			return true
		}
	}
	return false
}

// NormalizeAndAbsPath normalizes a path (using slashes) and converts it to an absolute path.
func NormalizeAndAbsPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return NormalizePath(absPath), nil
}

// NormalizePath ensures a consistent slash style for paths (uses `/`).
func NormalizePath(path string) string {
	return filepath.ToSlash(path)
}

// CreateTestFile creates temporary file with given content.
func CreateTestFile(filePath, content string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

// normalizeBaseURL ensures the base URL ends with a "/".
func NormalizeBaseURL(baseURL string) string {
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	return baseURL
}

// CountNonEmpty counts the number of non-empty strings in the provided list.
func CountNonEmpty(values ...string) int {
	count := 0
	for _, v := range values {
		if v != "" {
			count++
		}
	}
	return count
}
