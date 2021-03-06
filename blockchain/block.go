package blockchain

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"io"
	"math/big"
	"strings"
)

const (
	checksumLength = 4
	version        = byte(0x00)
)

// Block strcut
type Block struct {
	PrevHash     []byte
	Hash         []byte
	Nonce        int
	Signature    []byte
	Token        []byte
	PublicKey    []byte
	Transactions []*Transaction
}

// Sign signs block
func (block *Block) Sign(privateKey *ecdsa.PrivateKey) error {
	var data []byte
	for _, tx := range block.Transactions {
		data = append(data, tx.Data...)
	}

	hash := sha256.Sum256(data)

	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return err
	}

	block.Signature = append(r.Bytes(), s.Bytes()...)
	return nil
}

// VerifySignature verfies signature of the block
func (block *Block) VerifySignature() bool {
	r := big.Int{}
	s := big.Int{}
	r.SetBytes(block.Signature[:len(block.Signature)/2])
	s.SetBytes(block.Signature[len(block.Signature)/2:])

	x := big.Int{}
	y := big.Int{}
	keyLen := len(block.PublicKey)
	x.SetBytes(block.PublicKey[:(keyLen / 2)])
	y.SetBytes(block.PublicKey[(keyLen / 2):])

	rawPublicKey := ecdsa.PublicKey{elliptic.P256(), &x, &y}

	var data []byte
	for _, tx := range block.Transactions {
		data = append(data, tx.Data...)
	}
	hash := sha256.Sum256(data)

	validity := ecdsa.Verify(&rawPublicKey, hash[:], &r, &s)
	return validity
}

// HashTransactions hashes all transaction using merkle tree
func (block *Block) HashTransactions() []byte {
	var txHashes [][]byte
	for _, tx := range block.Transactions {
		txHashes = append(txHashes, tx.Serialize())
	}
	tree := NewMerkleTree(txHashes)
	return tree.RootNode.Data
}

// Encrypt encrypts all transaction
func (block *Block) Encrypt(passphrase []byte) {
	cipherBlock, _ := aes.NewCipher([]byte(Hash(passphrase)))
	gcm, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	for _, tx := range block.Transactions {
		tx.Data = gcm.Seal(nonce, nonce, tx.Data, nil)
	}
}

// Decrypt decrypts
func (block *Block) Decrypt(passphrase []byte) {
	key := []byte(Hash(passphrase))
	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()

	for _, tx := range block.Transactions {
		_, err := gcm.Open(nil, tx.Data[:nonceSize], tx.Data[nonceSize:], nil)
		if err != nil {
			panic(err)
		}
	}
}

// Serialize serialiszes block
func (block *Block) Serialize() ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(block)
	if err != nil {
		return []byte{}, err
	}
	return buffer.Bytes(), nil
}

// Deserialize deserializes block
func Deserialize(data []byte) (*Block, error) {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	if err != nil {
		return &block, err
	}
	return &block, nil
}

// Address function makes address from block token
func Address(token []byte) ([]byte, error) {
	pubHash, err := PublicKeyTokenHash(token)
	if err != nil {
		return []byte{}, err
	}
	versionedHash := append([]byte{version}, pubHash...)
	checksum := Checksum(versionedHash)

	fullHash := append(versionedHash, checksum...)
	address := Base58Encode(fullHash)

	return address, nil
}

// IsGenesis returns true if block is genesis, false otherwise
func (block *Block) IsGenesis() bool {
	if len(block.PrevHash) == 0 {
		return true
	}
	return false
}

// NewGenesisBlock crreates and returns a new genesis block
func NewGenesisBlock(token []byte, privateKey *ecdsa.PrivateKey) (*Block, error) {
	trans := Transaction{
		Data: []byte("Genesis Transaction"),
	}
	block := Block{
		Transactions: []*Transaction{&trans},
		Token:        token,
	}
	err := block.Sign(privateKey)
	if err != nil {
		return nil, err
	}
	pow := NewProof(&block)
	nonce, hash := pow.Run()
	block.Nonce = nonce
	block.Hash = hash

	return &block, nil
}

// String prints the block
func (block Block) String() string {
	var values []string

	values = append(values, fmt.Sprintf("\n----Block: "))
	values = append(values, fmt.Sprintf(" PrevHash  : %X", block.PrevHash))
	values = append(values, fmt.Sprintf(" Hash      : %X", block.Hash))
	values = append(values, fmt.Sprintf(" Nounce    : %d", block.Nonce))
	values = append(values, fmt.Sprintf(" Signature : %X", block.Signature))
	values = append(values, fmt.Sprintf(" Token     : %X", block.Token))
	values = append(values, fmt.Sprintf(" PublicKey : %X", block.PublicKey))

	// for idx := range block.Transactions {
	// 	values = append(values, fmt.Sprintf("   ├──Transaction  : %s", string(block.Transactions[idx].Data)))
	// }
	values = append(values, fmt.Sprintf(" Transactions(count): %v", len(block.Transactions)))
	return strings.Join(values, "\n")
}
