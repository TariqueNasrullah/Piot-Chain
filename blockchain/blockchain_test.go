package blockchain

import (
	"bytes"
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
