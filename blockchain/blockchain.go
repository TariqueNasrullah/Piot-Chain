package blockchain

import (
	"errors"
	"fmt"
	"time"

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
func (chain *BlockChain) AddBlock(block *Block) error {
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

	for {
		err := chain.Database.Update(func(txn *badger.Txn) error {
			address, err := Address(block.Token)
			if err != nil {
				return err
			}

			_, err = txn.Get(block.PrevHash)
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

			err = txn.Set(address, block.Hash)
			if err != nil {
				return err
			}

			return nil
		})

		if err == badger.ErrConflict {
			time.Sleep(time.Duration(time.Millisecond * 500))
			continue
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
		break
	}
	return nil
}

// AddGenesis adds genesis block to the chain
func (chain *BlockChain) AddGenesis(genesis *Block) error {
	if len(genesis.PrevHash) != 0 {
		return &ChainError{
			StatusCode: ErrorUnknown,
			Err:        errors.New("Block is not genesis"),
		}
	}
	pow := NewProof(genesis)
	valid := pow.Validate()
	if !valid {
		return &ChainError{
			StatusCode: ErrorInvalidProofOfWork,
			Err:        errors.New("Invalid proof of work"),
		}
	}

	for {
		err := chain.Database.Update(func(txn *badger.Txn) error {
			address, err := Address(genesis.Token)
			if err != nil {
				return err
			}

			_, err = txn.Get(address)
			if err == nil {
				return &ChainError{
					StatusCode: ErrorGenesisExists,
					Err:        errors.New("Genesis block exists"),
				}
			}

			data, err := genesis.Serialize()
			if err != nil {
				return err
			}

			err = txn.Set(genesis.Hash, data)
			if err != nil {
				return err
			}

			err = txn.Set(address, genesis.Hash)
			if err != nil {
				return err
			}

			return nil
		})

		if err == badger.ErrConflict {
			time.Sleep(time.Duration(time.Millisecond * 500))
			continue
		}

		if err != nil {
			return &ChainError{
				StatusCode: ErrorUnknown,
				Err:        fmt.Errorf("Error: %v", err),
			}
		}
		break
	}
	return nil
}
