package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

func AesEncrypt(key []byte, text string) (string, error) {
	textBytes := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	b := base64.StdEncoding.EncodeToString(textBytes)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func AesDecrypt(key []byte, text string) (string, error) {
	textBytes, err := base64.StdEncoding.DecodeString(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	if len(textBytes) < aes.BlockSize {
		return "", errors.New("Ciphertext too short")
	}
	iv := textBytes[:aes.BlockSize]
	textBytes = textBytes[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(textBytes, textBytes)
	data, err := base64.StdEncoding.DecodeString(string(textBytes))
	if err != nil {
		return "", err
	}
	return string(data), nil
}
