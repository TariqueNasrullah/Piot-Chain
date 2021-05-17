package blockchain

import (
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"os"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/sirupsen/logrus"
)

var (
	// DBPATH holds db path
	DBPATH = "tmp/database/"
)

// BlockChain structure
type BlockChain struct {
	Database *badger.DB
	repo     Repository
}

var (
	// Chain holds chain in memory
	Chain *BlockChain
)

// InitBlockChain initiates blockchain
func InitBlockChain(dbPath string, repo Repository) (*BlockChain, error) {
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	r := NewBadgerRepository(db)

	chain := BlockChain{Database: db, repo: r}

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
		logrus.Warn("pow is not valid")
		return &ChainError{
			StatusCode: ErrorInvalidProofOfWork,
			Err:        errors.New("Invalid proof of work"),
		}
	}
	for {
		err := chain.Database.View(func(txn *badger.Txn) error {
			if _, err := txn.Get(block.Hash); err == badger.ErrKeyNotFound {
				return nil
			}

			return errors.New("Key Exists")
		})
		if err != nil {
			return err
		}
		err = chain.Database.Update(func(txn *badger.Txn) error {
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
	logrus.Infoln("Added block to local chain")
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

	address, err := Address(genesis.Token)
	if err != nil {
		return err
	}

	_, err = chain.repo.GetTail(context.Background(), string(address))
	if err == nil {
		return &ChainError{
			StatusCode: ErrorGenesisExists,
			Err:        errors.New("Genesis block exists"),
		}
	}

	err = chain.repo.Store(context.Background(), string(address), genesis)

	if err != nil {
		return &ChainError{
			StatusCode: ErrorUnknown,
			Err:        fmt.Errorf("Error: %v", err),
		}
	}

	return nil
}

// FullHeight returns blockchain height
func (chain *BlockChain) FullHeight() int64 {
	height, _ := chain.repo.CollectionCount(context.Background())
	return height
}

// Height retunrs height of a chain
func (chain *BlockChain) Height(token []byte) (int64, error) {
	addr, err := Address(token)
	if err != nil {
		return 0, err
	}

	blockList, err := chain.repo.Fetch(context.Background(), string(addr))
	if err != nil {
		return 0, err
	}

	return int64(len(blockList)), nil
}

// Chain retunrs chain of a token
func (chain *BlockChain) Chain(token []byte) ([]*Block, error) {
	addr, err := Address(token)
	if err != nil {
		return nil, err
	}

	blockList, err := chain.repo.Fetch(context.Background(), string(addr))
	if err != nil {
		return nil, err
	}

	return blockList, nil
}

// LastHash returns last hash
func (chain *BlockChain) LastHash(token []byte) ([]byte, error) {
	addr, err := Address(token)
	if err != nil {
		return []byte{}, err
	}

	return chain.repo.GetTail(context.Background(), string(addr))
}

// ClearDB clear Blockchain Database
func ClearDB() error {
	err := os.Remove(DBPATH + "MANIFEST")
	if err != nil {
		return err
	}
	return nil
}
