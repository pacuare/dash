package shared

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"log"
	"os"
)

var (
	AES cipher.Block
	GCM cipher.AEAD
)

func init() {
	var err error
	AES, err = aes.NewCipher([]byte(os.Getenv("SECRET_KEY_BASE")))

	if err != nil {
		log.Fatalf("Failed to create secret key: %e", err)
	}

	GCM, err = cipher.NewGCM(AES)

	if err != nil {
		log.Fatalf("Failed to create GCM: %e", err)
	}
}

func Encrypt(data []byte) ([]byte, error) {
	nonce := make([]byte, GCM.NonceSize())

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return GCM.Seal(nonce, nonce, data, nil), nil
}

func Decrypt(data []byte) ([]byte, error) {
	nonceSize := GCM.NonceSize()

	if len(data) < nonceSize {
		return nil, errors.New("data smaller than nonce")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	return GCM.Open(nil, nonce, ciphertext, nil)
}
