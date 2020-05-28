package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
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

/ Encrypt encrypts all transaction
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