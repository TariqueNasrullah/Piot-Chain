package blockchain

import "github.com/dgraph-io/badger"

// BlockChain structure
type BlockChain struct {
	Database *badger.DB
}
