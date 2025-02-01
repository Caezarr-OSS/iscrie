package scripts

import (
	"iscrie/utils"
	"os"
	"path/filepath"
)

// CleanupTestData recursively deletes all files and directories under the specified root path.
func CleanupTestData(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return utils.LogAndReturnError("Failed to access path %s: %v", path, err)
		}

		if info.IsDir() && path != root {
			// Remove directory
			if err := os.RemoveAll(path); err != nil {
				return utils.LogAndReturnError("Failed to remove directory %s: %v", path, err)
			}
			utils.LogInfo("Removed directory: %s", path)
		} else if !info.IsDir() {
			// Remove file
			if err := os.Remove(path); err != nil {
				return utils.LogAndReturnError("Failed to remove file %s: %v", path, err)
			}
			utils.LogInfo("Removed file: %s", path)
		}

		return nil
	})
}
