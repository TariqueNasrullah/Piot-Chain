package blockchain

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"log"
)

// Hash hashes
func Hash(key []byte) string {
	hasher := md5.New()
	hasher.Write(key)
	return hex.EncodeToString(hasher.Sum(nil))
}

// ToHex converts int to bytes
func ToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)

	}

	return buff.Bytes()
}
