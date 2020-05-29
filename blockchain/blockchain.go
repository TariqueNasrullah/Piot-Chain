package blockchain

import (
	"errors"
	"fmt"

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

// AddBlock adds a block to the chain
// 1. chekcs signature
// 2. checks pow
// 3. checks existance of previous hash
func (chain *BlockChain) AddBlock(block *Block) *ChainError {
	valid := block.VerifySignature()
	if !valid {
		return &ChainError{
			StatusCode: ErrorInvalidSignature,
			Err:        errors.New("Signature can't be verified"),
		}
	}

	pow := NewProof(block)
	valid = pow.Validate()
	if !valid {
		return &ChainError{
			StatusCode: ErrorInvalidProofOfWork,
			Err:        errors.New("Invalid proof of work"),
		}
	}

UpdateDB:
	err := chain.Database.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(block.PrevHash)
		if err == badger.ErrKeyNotFound {
			return err
		}

		data, err := block.Serialize()
		if err != nil {
			return err
		}

		err = txn.Set(block.Hash, data)
		if err != nil {
			return err
		}

		address, err := block.Address()
		if err != nil {
			return err
		}

		err = txn.Set(address, block.Hash)
		if err != nil {
			return err
		}

		return nil
	})

	if err == badger.ErrConflict {
		goto UpdateDB
	}
	if err == badger.ErrKeyNotFound {
		return &ChainError{
			StatusCode: ErrorPreviousHashNotFound,
			Err:        errors.New("Previous hash not found"),
		}
	}
	if err != nil {
		return &ChainError{
			StatusCode: ErrorUnknown,
			Err:        fmt.Errorf("Error: %v", err),
		}
	}
	return nil
}

func test() {
	fmt.Println("testing")
}
