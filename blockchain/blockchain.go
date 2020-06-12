package blockchain

import (
	"errors"
	"fmt"
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
}

//Iterator structure
type Iterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

var (
	// Chain holds chain in memory
	Chain *BlockChain
)

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

// FullHeight returns blockchain height
func (chain *BlockChain) FullHeight() int64 {
	height := int64(0)

	chain.Database.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			height++
		}
		return nil
	})
	return height
}

// Height retunrs height of a chain
func (chain *BlockChain) Height(token []byte) (int64, error) {
	height := int64(0)
	addr, err := Address(token)
	if err != nil {
		return 0, err
	}

	keyfound := true
	var lastHash []byte

	err = chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(addr)
		if err == badger.ErrKeyNotFound {
			keyfound = false
			return nil
		}
		err = item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	if keyfound == false {
		return 0, nil
	}

	itr := Iterator{CurrentHash: lastHash, Database: chain.Database}

	for {
		block := itr.Next()
		if block == nil {
			break
		}

		height++

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return height, nil
}

// Chain retunrs chain of a token
func (chain *BlockChain) Chain(token []byte) ([]*Block, error) {
	blockList := []*Block{}
	addr, err := Address(token)
	if err != nil {
		return blockList, err
	}

	keyfound := true
	var lastHash []byte

	err = chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(addr)
		if err == badger.ErrKeyNotFound {
			keyfound = false
			return nil
		}
		err = item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return blockList, err
	}
	if keyfound == false {
		return blockList, nil
	}

	itr := Iterator{CurrentHash: lastHash, Database: chain.Database}

	for {
		block := itr.Next()
		if block == nil {
			break
		}

		blockList = append(blockList, block)

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return blockList, nil
}

// LastHash returns last hash
func (chain *BlockChain) LastHash(token []byte) ([]byte, error) {
	addr, err := Address(token)
	if err != nil {
		return []byte{}, err
	}
	var lastHash []byte
	err = chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(addr)
		if err == badger.ErrKeyNotFound {
			return errors.New("Last Hash not found, you need to sync your localchain from a miner")
		}
		err = item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return []byte{}, err
	}
	return lastHash, nil
}

// ClearDB clear Blockchain Database
func ClearDB() error {
	err := os.Remove(DBPATH + "MANIFEST")
	if err != nil {
		return err
	}
	return nil
}

// Next returns next block
func (itr *Iterator) Next() *Block {
	var block *Block

	err := itr.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(itr.CurrentHash)
		if err != nil {
			return err
		}
		var encodedBlock []byte
		err = item.Value(func(val []byte) error {
			encodedBlock = val
			return nil
		})
		if err != nil {
			return err
		}
		block, err = Deserialize(encodedBlock)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil
	}
	itr.CurrentHash = block.PrevHash
	return block
}
