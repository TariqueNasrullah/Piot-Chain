package blockchain

import (
	"crypto/md5"
	"encoding/hex"
)

// Hash hashes
func Hash(key []byte) string {
	hasher := md5.New()
	hasher.Write(key)
	return hex.EncodeToString(hasher.Sum(nil))
}
