package blockchain

import (
	"fmt"
)

const (
	// ErrorInvalidSignature status code
	ErrorInvalidSignature = 401
	// ErrorInvalidProofOfWork status code
	ErrorInvalidProofOfWork = 402
	// ErrorPreviousHashNotFound status code
	ErrorPreviousHashNotFound = 403
	// ErrorGenesisExists status code
	ErrorGenesisExists = 404
	// ErrorUnknown status code
	ErrorUnknown = 420
)

// ChainError is custom error structure
type ChainError struct {
	StatusCode int
	Err        error
}

// Error stringify ChainError structure
func (cErr *ChainError) Error() string {
	return fmt.Sprintf("status %d: Error %v", cErr.StatusCode, cErr.Err)
}
