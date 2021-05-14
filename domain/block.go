package domain

import "context"

// Block struct
type Block struct {
	PrevHash     []byte
	Hash         []byte
	Nonce        int
	Signature    []byte
	Token        []byte
	PublicKey    []byte
	Transactions []*Transaction
}

type BlockRepository interface {
	Fetch(ctx context.Context) ([]*Block, error)
	GetByID(ctx context.Context, id string) (*Block, error)
	Store(ctx context.Context, subject *Block) error
	Delete(ctx context.Context, id string) error
}

type BlockUseCase interface {
	Fetch(ctx context.Context) ([]*Block, error)
	GetByID(ctx context.Context, id string) (*Block, error)
	Store(ctx context.Context, subject *Block) error
	Delete(ctx context.Context, id string) error
}
