package importer

import (
	"fmt"
	"iscrie/utils"
	"sync"
)

// BatchError represents a structured error for batch processing.
type BatchError struct {
	Errors map[interface{}]error
}

// NewBatchError creates a new BatchError instance and logs it.
func NewBatchError(errors map[interface{}]error) *BatchError {
	formattedMessage := FormatBatchErrorMessage(errors)
	utils.LogError(formattedMessage)
	return &BatchError{Errors: errors}
}

// FormatBatchErrorMessage formats the error message for batch processing failures.
func FormatBatchErrorMessage(errors map[interface{}]error) string {
	return fmt.Sprintf("Batch processing failed for %d items", len(errors))
}

// Error implements the error interface for BatchError.
func (e *BatchError) Error() string {
	return FormatBatchErrorMessage(e.Errors)
}

// ProcessBatch processes items in batches with a given batch size and processing function.
func ProcessBatch[T any](items []T, batchSize int, process func(T) error) error {
	if batchSize <= 0 {
		batchSize = 1
	}

	utils.LogDebug("Starting batch processing with batch size: %d", batchSize)

	semaphore := make(chan struct{}, batchSize)
	errChan := make(chan struct {
		Item T
		Err  error
	}, len(items))

	var wg sync.WaitGroup

	for _, item := range items {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(i T) {
			defer func() {
				<-semaphore
				wg.Done()
			}()
			if err := process(i); err != nil {
				utils.LogError("Error processing item: %v, error: %v", i, err)
				errChan <- struct {
					Item T
					Err  error
				}{Item: i, Err: err}
			}
		}(item)
	}

	// Wait for all tasks to complete
	wg.Wait()
	close(errChan)

	// Collect errors
	errors := make(map[interface{}]error)
	for result := range errChan {
		errors[result.Item] = result.Err
	}
	if len(errors) > 0 {
		utils.LogError("Batch processing encountered %d errors.", len(errors))
		return NewBatchError(errors)
	}

	utils.LogDebug("Batch processing completed successfully.")
	return nil
}
