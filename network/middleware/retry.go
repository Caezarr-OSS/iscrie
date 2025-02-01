package middleware

import (
	"iscrie/utils"
	"time"
)

// Retry retries the provided operation up to a given number of attempts, with a specified delay between attempts.
// It supports exponential backoff and logs errors if a logger is provided.
func Retry(attempts int, delay time.Duration, operation func() error) error {
	if attempts <= 0 {
		return utils.LogAndReturnError("Attempts must be greater than 0")
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if lastErr = operation(); lastErr == nil {
			utils.LogDebug("Operation succeeded on attempt %d.", attempt)
			return nil
		}

		utils.LogError("Attempt %d/%d failed: %v", attempt, attempts, lastErr)
		if attempt < attempts {
			utils.LogDebug("Retrying in %s...", delay)
			time.Sleep(delay)
			delay *= 2 // Exponential backoff
		}
	}

	return utils.LogAndReturnError("Operation failed after %d attempts: %w", attempts, lastErr)
}
