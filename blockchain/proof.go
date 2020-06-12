package blockchain

import (
	"bytes"
	"crypto/sha256"
	"math"
	"math/big"
)

var (
	// Difficulty is POW difficulty
	Difficulty = 12
)

// ProofOfWork struct
type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

// NewProof returns new ProofOfWork
func NewProof(b *Block) *ProofOfWork {
	target := big.NewInt(1)

	target.Lsh(target, uint(256-Difficulty))

	pow := &ProofOfWork{Block: b, Target: target}
	return pow
}

// InitData initiates data
func (pow *ProofOfWork) InitData(nonce int) []byte {
	var tmp []byte
	for idx := range pow.Block.Transactions {
		tmp = append(tmp, pow.Block.Transactions[idx].Data...)
	}
	trsHash := sha256.Sum256(tmp)
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevHash,
			trsHash[:],
			ToHex(int64(nonce)),
			ToHex(int64(Difficulty)),
		},
		[]byte{},
	)
	return data
}

// Run proof of work
func (pow *ProofOfWork) Run() (int, []byte) {
	var intHash big.Int
	var hash [32]byte

	nonce := 0

	for nonce < math.MaxInt64 {
		data := pow.InitData(nonce)
		hash = sha256.Sum256(data)

		intHash.SetBytes(hash[:])

		if intHash.Cmp(pow.Target) == -1 {
			break
		} else {
			nonce++
		}
	}
	return nonce, hash[:]
}

// Validate validates the proof of work
func (pow *ProofOfWork) Validate() bool {
	var intHash big.Int
	data := pow.InitData(pow.Block.Nonce)
	hash := sha256.Sum256(data)
	intHash.SetBytes(hash[:])
	return intHash.Cmp(pow.Target) == -1
}
