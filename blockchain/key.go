package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Key structure
type Key struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  []byte
	SecretKey  []byte
	Token      []byte
}

var (
	// KEYPATH is the path of key file
	KEYPATH = "tmp/key/key.data"
)

// GenerateKey generates SecretKey, privatekey, publickey
func GenerateKey(keyPath string) (*Key, error) {
	var key Key

	// Assign SecretKey
	var randomKey [32]byte
	rand.Read(randomKey[:])
	key.SecretKey = randomKey[:]

	// Gnerate ecdsa priv pub key
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return &key, err
	}
	key.PrivateKey = privKey
	key.PublicKey = append(privKey.PublicKey.X.Bytes(), privKey.PublicKey.Y.Bytes()...)

	return &key, nil
}

// LoadKey loads key from file
func LoadKey(keyPath string) (*Key, error) {
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return &Key{}, errors.New("Key File Does not exist")
	}

	var key Key

	fileContent, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return &key, err
	}

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&key)
	if err != nil {
		return &key, err
	}

	return &key, nil
}

// SaveFile saves key into file
func (key *Key) SaveFile(keyPath string) error {
	var buffer bytes.Buffer
	gob.Register(elliptic.P256())

	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(key)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(keyPath, buffer.Bytes(), 0644)
	if err != nil {
		return err
	}
	return nil
}

func (key *Key) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf(" --------- Device Key Information:"))
	lines = append(lines, fmt.Sprintf(" PubKey     : %x", key.PublicKey))
	lines = append(lines, fmt.Sprintf(" SecretKey  : %x", key.SecretKey))
	lines = append(lines, fmt.Sprintf(" Token      : %x", key.Token))

	return strings.Join(lines, "\n")
}
