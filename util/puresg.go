package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

const privateKey = "1b24b23172ebc8aae4750b41e27f91e355dca107cc377addc27cea5ddb4036ac"

func EncryptPureSGSecret(plaintext string) (string, error) {

	key, err := hex.DecodeString(privateKey)
	if err != nil {
		fmt.Println("Error decoding key:", err)
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	return hex.EncodeToString(nonce) + hex.EncodeToString(ciphertext), nil
}

func DecryptPureSGSecret(encrypted string) (string, error) {
	key, err := hex.DecodeString(privateKey)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	// Calculate the minimum length: nonce + ciphertext
	if len(encrypted) < nonceSize*2 {
		return "", errors.New("encrypted data too short")
	}

	// Decode the nonce (first part of the string)
	nonce, err := hex.DecodeString(encrypted[:nonceSize*2])
	if err != nil {
		return "", err
	}

	// Decode the ciphertext (rest of the string)
	ciphertext, err := hex.DecodeString(encrypted[nonceSize*2:])
	if err != nil {
		return "", err
	}

	// Decrypt using AES-GCM
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
