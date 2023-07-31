package util

import (
	"crypto/aes"
	"crypto/cipher"
	random "crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"
)

// RandStringRunes 返回随机字符串
func RandStringRunes(n int) string {
	var letterRunes = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func AesEncrypt(str []byte) string {
	key := []byte(os.Getenv("ENCRYPT_KEY"))
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println(err)
	}
	ciphertext := make([]byte, aes.BlockSize+len(str))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(random.Reader, iv); err != nil {
		fmt.Println(err)
	}
	mode := cipher.NewCFBEncrypter(block, iv)
	mode.XORKeyStream(ciphertext[aes.BlockSize:], str)
	return base64.RawStdEncoding.EncodeToString(ciphertext)
}