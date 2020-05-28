package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"

	"github.com/TariqueNasrullah/Private-IoT-blockchain/blockchain"
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

func main() {
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

	fmt.Printf("%s\n", block)

	validSig := block.VerifySignature()
	fmt.Printf("Signature Verification: %v\n", validSig)
}
