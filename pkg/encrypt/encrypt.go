package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
)

func Encrypt(value, salt string) (chipertext string, err error) {
	block, err := aes.NewCipher([]byte(salt))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return
	}
	ciphertextByte := gcm.Seal(
		nonce,
		nonce,
		[]byte(value),
		nil)
	chipertext = base64.StdEncoding.EncodeToString(ciphertextByte)

	return
}

func Decrypt(cipherText, key string) (plainText string, err error) {
	// prepare cipher
	keyByte := []byte(key)
	block, err := aes.NewCipher(keyByte)
	if err != nil {
		return
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}
	nonceSize := gcm.NonceSize()

	ciphertextByte, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}
	nonce, ciphertextByteClean := ciphertextByte[:nonceSize], ciphertextByte[nonceSize:]
	plaintextByte, err := gcm.Open(
		nil,
		nonce,
		ciphertextByteClean,
		nil)
	if err != nil {
		log.Println(err)
		return
	}
	plainText = string(plaintextByte)

	return
}
