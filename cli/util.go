package cli

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/TariqueNasrullah/iotchain/blockchain"
)

var (
	keyPath = "key.data"
)

func generateToken(username, password string) ([]byte, error) {
	hash := []byte(username + password)

	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	r, s, err := ecdsa.Sign(rand.Reader, key, hash[:])
	if err != nil {
		return []byte{}, err
	}

	signature := append(r.Bytes(), s.Bytes()...)
	return signature, nil
}
func handle(err error) {
	if err != nil {
		panic(err)
	}
}

func populateDb() error {
	token, err := generateToken("admin", "pass")
	handle(err)
	key, err := blockchain.GenerateKey(keyPath)
	handle(err)
	key.Token = token

	trans := blockchain.Transaction{
		Data: []byte("Hello Transaction"),
	}
	block := blockchain.Block{
		Transactions: []*blockchain.Transaction{&trans},
		Token:        key.Token,
		PublicKey:    key.PublicKey,
	}

	err = block.Sign(key.PrivateKey)
	handle(err)

	pow := blockchain.NewProof(&block)
	nonce, hash := pow.Run()
	block.Nonce = nonce
	block.Hash = hash

	err = blockchain.Chain.AddGenesis(&block)
	if err != nil {
		return err
	}
	return nil
}
