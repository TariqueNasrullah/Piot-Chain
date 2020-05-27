package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
)

// Transaction struct
type Transaction struct {
	Data []byte
}

// Serialize sereilizes transactions
func (tx *Transaction) Serialize() []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return buffer.Bytes()
}
