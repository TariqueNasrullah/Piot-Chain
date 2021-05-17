package blockchain

import (
	"context"
	"github.com/dgraph-io/badger"
)

type Repository interface {
	Fetch(ctx context.Context, address string) ([]*Block, error)
	Store(ctx context.Context, address string, block *Block) error
	GetTail(ctx context.Context, address string) ([]byte, error)
	CollectionCount(ctx context.Context) (int64, error)
}

type repository struct {
	db *badger.DB
}

func NewBadgerRepository(db *badger.DB) Repository {
	return &repository{db: db}
}

func (r repository) Fetch(ctx context.Context, address string) ([]*Block, error) {
	var blockList []*Block

	tail, err := r.GetTail(ctx, address)
	if err != nil {
		return nil, err
	}

	for {
		var block *Block
		err := r.db.View(func(txn *badger.Txn) error {
			item, err := txn.Get(tail)
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
			break
		}
		blockList = append(blockList, block)
		tail = block.Hash
	}
	return blockList, nil
}

func (r repository) Store(ctx context.Context, address string, block *Block) error {
	addr := []byte(address)

	data, err := block.Serialize()
	if err != nil {
		return err
	}

	err = r.db.Update(func(txn *badger.Txn) error {
		err = txn.Set(block.Hash, data)
		if err != nil {
			return err
		}

		return txn.Set(addr, block.Hash)
	})

	return err
}

func (r repository) GetTail(ctx context.Context, address string) ([]byte, error) {
	var byteBlock []byte
	addr := []byte(address)
	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(addr)
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			byteBlock = val
			return nil
		})
		return err
	})
	return byteBlock, err
}

func (r repository) CollectionCount(ctx context.Context) (int64, error) {
	height := int64(0)

	err := r.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			height++
		}
		return nil
	})
	return height, err
}
