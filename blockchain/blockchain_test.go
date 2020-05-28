package blockchain

import (
	"bytes"
	"fmt"
	"os"
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
