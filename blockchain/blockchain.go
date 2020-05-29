package blockchain

import (
	"github.com/dgraph-io/badger"
)

// BlockChain structure
type BlockChain struct {
	Database *badger.DB
}

// InitBlockChain initiates blockchain
func InitBlockChain(dbPath string) (*BlockChain, error) {
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	chain := BlockChain{Database: db}

	return &chain, nil
}
