package blockchain

import (
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/dgraph-io/badger"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

//go:generate protoc -I=. --go_out=plugins=grpc:. miner.proto

// Server structure
type Server struct {
	UnimplementedMinerServer
}

// Test tests
func (srv *Server) Test(ctx context.Context, in *TestRequest) (*TestResponse, error) {
	trans := Transaction{Data: []byte("data")}
	block := Block{
		PrevHash:     []byte("hash"),
		Transactions: []*Transaction{&trans},
		Token:        []byte("token"),
		PublicKey:    []byte("pubkey"),
	}

	pow := NewProof(&block)
	nonce, hash := pow.Run()
	block.Nonce = nonce
	block.Hash = hash

	validate := pow.Validate()
	if !validate {
		logrus.Error("sending block invalid")
	} else {
		logrus.Info("sending block valid")
	}
	serializedBlock, err := block.Serialize()
	if err != nil {
		logrus.Fatal(err)
	}
	fmt.Printf("%s\n", block)
	return &TestResponse{Block: serializedBlock}, nil
}

// SendAddress implementation
func (srv *Server) SendAddress(ctx context.Context, in *SendAddressRequest) (*SendAddressResponse, error) {
	KnownNodes[in.Addr] = struct{}{}

	if _, ok := ConnectedNodes[in.Addr]; !ok {
		network := Network{}
		go network.SendAddress(in.Addr)
	}
	return &SendAddressResponse{ResponseText: "OK", StatusCode: 200}, nil
}

// GetAddress returns a stream of address that this node is connected to
func (srv *Server) GetAddress(in *GetAddressRequest, stream Miner_GetAddressServer) error {
	for addr := range ConnectedNodes {
		if err := stream.Send(&GetAddressResponse{Address: addr}); err != nil {
			return err
		}
	}
	return nil
}

// FullHeight returns blockchain fullheight
func (srv *Server) FullHeight(context.Context, *FullHeightRequest) (*FullHeightResponse, error) {
	height := Chain.FullHeight()
	return &FullHeightResponse{Height: height}, nil
}

// GetFullChain streams back the full blockchain
func (srv *Server) GetFullChain(in *GetFullChainRequest, stream Miner_GetFullChainServer) error {
	err := Chain.Database.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			var value []byte

			err := item.Value(func(val []byte) error {
				value = val
				return nil
			})
			if err != nil {
				return err
			}

			if err := stream.Send(&GetFullChainResponse{Key: key, Value: value}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// GetChain returns stream of block
func (srv *Server) GetChain(in *GetChainRequest, stream Miner_GetChainServer) error {
	blockList, err := Chain.Chain(in.Token)
	if err != nil {
		return err
	}
	for i := len(blockList) - 1; i >= 0; i-- {
		serializedBlock, err := blockList[i].Serialize()
		if err != nil {
			return err
		}
		stream.Send(&GetChainResponse{Block: serializedBlock})
	}
	return nil
}

// PropagateBlock propagates a block accross the network
func (srv *Server) PropagateBlock(ctx context.Context, in *PropagateBlockRequest) (*PropagateBlockResponse, error) {
	block, err := Deserialize(in.Block)
	if err != nil {
		return nil, err
	}

	if block.IsGenesis() {
		err := Chain.AddGenesis(block)
		if err != nil {
			return nil, err
		}
	} else {
		err = Chain.AddBlock(block)
		if err != nil {
			return nil, err
		}
	}

	network := Network{}
	for addr := range ConnectedNodes {
		go network.PropagateBlock(in.Block, addr)
	}

	return &PropagateBlockResponse{Ok: true}, nil
}

// Token returns a token
func (srv *Server) Token(ctx context.Context, in *TokenRequest) (*TokenResponse, error) {
	if in.Username == "" || in.Password == "" {
		return nil, errors.New("Invalid Username or Password")
	}
	hash := []byte(in.Username + in.Password)
	key, err := LoadKey(KEYPATH)
	if err != nil {
		return nil, err
	}
	r, s, err := ecdsa.Sign(rand.Reader, key.PrivateKey, hash[:])
	if err != nil {
		return nil, err
	}
	signature := append(r.Bytes(), s.Bytes()...)

	block, err := NewGenesisBlock(signature, key.PrivateKey)
	if err != nil {
		return nil, err
	}
	serializedBlock, err := block.Serialize()
	if err != nil {
		return nil, err
	}

	err = Chain.AddGenesis(block)
	if err != nil {
		return nil, err
	}

	network := Network{}
	for addr := range ConnectedNodes {
		go network.PropagateBlock(serializedBlock, addr)
	}

	return &TokenResponse{Token: signature}, nil
}

// Ping returns a simple response
func (srv *Server) Ping(ctx context.Context, in *PingRequest) (*PingResponse, error) {
	return &PingResponse{}, nil
}

// Height returns height of chain of token
func (srv *Server) Height(ctx context.Context, in *HeightRequest) (*HeightResponse, error) {
	height, err := Chain.Height(in.Token)
	if err != nil {
		return nil, err
	}
	return &HeightResponse{Height: height}, nil
}

// Mine mines
func (srv *Server) Mine(ctx context.Context, in *MineRequest) (*MineResponse, error) {
	block, err := Deserialize(in.Block)
	if err != nil {
		return nil, err
	}
	pow := NewProof(block)

	nonce, hash := pow.Run()
	block.Nonce = nonce
	block.Hash = hash

	serializedBlock, err := block.Serialize()
	if err != nil {
		return nil, err
	}

	err = Chain.AddBlock(block)
	if err != nil {
		return nil, err
	}

	network := Network{}
	for addr := range ConnectedNodes {
		go network.PropagateBlock(serializedBlock, addr)
	}

	return &MineResponse{Block: serializedBlock}, nil
}

// PrintConnectedNodes prints connected nodes
func PrintConnectedNodes() {
	fmt.Println("  --Connected Nodes")
	for key := range ConnectedNodes {
		fmt.Println("   ", key)
	}
}
