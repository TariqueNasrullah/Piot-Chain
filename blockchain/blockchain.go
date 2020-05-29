package blockchain

import (
	"os"

	"github.com/dgraph-io/badger"
)

// BlockChain structure
type BlockChain struct {
	Database *badger.DB
}

// DBexists checks if database file already exists
func DBexists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}
