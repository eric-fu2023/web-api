package util

import (
	"crypto/aes"
	"crypto/cipher"
	random "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"
)

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

func MapSlice[T any, M any](a []T, f func(T) M) []M {
	n := make([]M, len(a))
	for i, e := range a {
		n[i] = f(e)
	}
	return n
}

func JSON(jsonObj any) string {
	bytes, _ := json.Marshal(jsonObj)
	return string(bytes)
}

func Reduce[T, M any](s []T, f func(M, T) M, initValue M) M {
    acc := initValue
    for _, v := range s {
        acc = f(acc, v)
    }
    return acc
}