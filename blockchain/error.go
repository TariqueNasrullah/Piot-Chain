package blockchain

import "fmt"

// ChainError is custom error structure
type ChainError struct {
	StatusCode int
	Err        error
}

// Error stringify ChainError structure
func (cErr *ChainError) Error() string {
	return fmt.Sprintf("status %d: Error %v", cErr.StatusCode, cErr.Err)
}
