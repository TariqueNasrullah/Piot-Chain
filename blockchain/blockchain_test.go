package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestSerializeDeserialize(t *testing.T) {
	trans := Transaction{
		Data: []byte("Hello World"),
	}

	block := Block{
		Transactions: []*Transaction{&trans},
	}

	serializedData, err := block.Serialize()
	if err != nil {
		t.Fatalf("Error Not expected! Error: %v\n", err)
	}

	blockPrime, err := Deserialize(serializedData)
	if err != nil {
		t.Fatalf("Error Not expected! Error: %v\n", err)
	}

	// Compare two struct block and blockPrime
	if len(block.Transactions) != len(blockPrime.Transactions) {
		t.Fatal("Two block are not same")
	}

	if !bytes.Equal(block.Transactions[0].Data, blockPrime.Transactions[0].Data) {
		t.Fatal("Two block are not same")
	}
}

func TestKey(t *testing.T) {
	keyPath := "key.data"

	defer func() {
		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			return
		}
		os.Remove(keyPath)
	}()

	key, err := GenerateKey(keyPath)
	if err != nil {
		t.Fatalf("Error Not expected! Error: %v\n", err)
	}

	err = key.SaveFile(keyPath)
	if err != nil {
		t.Fatalf("Error Not expected! Error: %v\n", err)
	}
}

func TestProof(t *testing.T) {
	trans := Transaction{
		Data: []byte("Hello World"),
	}

	block := Block{
		Transactions: []*Transaction{&trans},
	}
	pow := NewProof(&block)
	nonce, hash := pow.Run()
	block.Nonce = nonce
	block.Hash = hash

	valid := pow.Validate()

	if !valid {
		t.FailNow()
	}
	fmt.Printf("POW Validity: %v Expected: true\n", valid)

	nonce, hash = pow.Run()
	block.Nonce = nonce - 1
	block.Hash = hash
	valid = pow.Validate()

	if valid {
		t.FailNow()
	}
	fmt.Printf("POW Validity: %v Expected: false\n", valid)
}

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

func TestAddress(t *testing.T) {
	token, err := generateToken("admin", "pass")
	if err != nil {
		t.Fatalf("Error Not expected! Error: %v\n", err)
	}
	key, err := GenerateKey("key.data")
	if err != nil {
		t.Fatalf("Error Not expected! Error: %v\n", err)
	}

	key.Token = token

	trans := Transaction{
		Data: []byte("Hello Transaction"),
	}
	block := Block{
		Transactions: []*Transaction{&trans},
		Token:        key.Token,
		PublicKey:    key.PublicKey,
	}

	addr0, err := Address(block.Token)
	if err != nil {
		t.Fatalf("Error Not expected! Error: %v\n", err)
	}

	addr1, err := Address(block.Token)
	if err != nil {
		t.Fatalf("Error Not expected! Error: %v\n", err)
	}

	if bytes.Compare(addr0, addr1) != 0 {
		t.Fatal("Two address should be same. Failed")
	}
}
func ensureDir(fileName string) error {
	dirName := filepath.Dir(fileName)
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if merr != nil {
			return merr
		}
	}
	return nil
}

func TestAddBlock(t *testing.T) {
	err := ensureDir("tmp")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.RemoveAll("tmp")
	}()

	token, err := generateToken("admin", "pass")
	if err != nil {
		t.Fatalf("Error Not expected! Error: %v\n", err)
	}
	key, err := GenerateKey("key.data")
	if err != nil {
		t.Fatalf("Error Not expected! Error: %v\n", err)
	}

	key.Token = token

	trans := Transaction{
		Data: []byte("Hello Transaction"),
	}
	genesisTransaction := Transaction{
		Data: []byte("Genesis Transaction"),
	}

	genesisBlock := Block{
		Transactions: []*Transaction{&genesisTransaction},
		Token:        key.Token,
	}
	pow := NewProof(&genesisBlock)
	nonce, hash := pow.Run()
	genesisBlock.Nonce = nonce
	genesisBlock.Hash = hash

	block := Block{
		Transactions: []*Transaction{&trans},
		Token:        key.Token,
		PublicKey:    key.PublicKey,
		PrevHash:     hash,
	}
	block.Sign(key.PrivateKey)
	pow = NewProof(&block)
	nonce, hash = pow.Run()
	block.Nonce = nonce
	block.Hash = hash

	chain, err := InitBlockChain("tmp")
	if err != nil {
		t.Fatal("Init chain Error is unexpected ", err)
	}
	defer chain.Database.Close()

	err = chain.AddGenesis(&genesisBlock)
	if err != nil {
		t.Fatalf("Error is unexpected: %v\n", err.Error())
	}

	err = chain.AddBlock(&block)

	if err != nil {
		t.Fatalf("Error is unexpected: %v\n", err.Error())
	}
}
