package security

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"os"
)

var secretKey []byte

func LoadKeyFromFile(path string) error {

	file, err := os.Open(path)

	if err != nil {

		return err
	}

	defer func(file *os.File) {

		err := file.Close()

		if err != nil {

			log.Printf("Error closing file: %v\n", err)
		}

	}(file)

	var config struct {
		Secret string `json:"encryption_secret"`
	}

	if err := json.NewDecoder(file).Decode(&config); err != nil {

		return err
	}

	key, err := base64.StdEncoding.DecodeString(config.Secret)

	if err != nil {

		return err
	}

	secretKey = key

	return nil
}

func Encrypt(plainText []byte) (string, error) {

	defer func() {

		if r := recover(); r != nil {

			log.Printf("Recovered from panic in Encrypt: %v\n", r)

		}
	}()

	block, err := aes.NewCipher(secretKey)

	if err != nil {

		return "", err
	}

	iv := make([]byte, aes.BlockSize)

	if _, err := rand.Read(iv); err != nil {

		return "", err
	}

	padded := pkcs5Pad(plainText, aes.BlockSize)

	mode := cipher.NewCBCEncrypter(block, iv)

	encrypted := make([]byte, len(padded))

	mode.CryptBlocks(encrypted, padded)

	output := append(iv, encrypted...)

	return base64.StdEncoding.EncodeToString(output), nil
}

func Decrypt(base64Cipher string) ([]byte, error) {

	defer func() {

		if r := recover(); r != nil {

			log.Printf("Recovered from panic in Decrypt: %v\n", r)
		}

	}()

	data, err := base64.StdEncoding.DecodeString(base64Cipher)

	if err != nil {

		return nil, err
	}

	if len(data) < aes.BlockSize {

		return nil, errors.New("ciphertext too short")
	}

	iv := data[:aes.BlockSize]

	ciphertext := data[aes.BlockSize:]

	block, err := aes.NewCipher(secretKey)

	if err != nil {

		return nil, err
	}

	if len(ciphertext)%aes.BlockSize != 0 {

		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	decrypted := make([]byte, len(ciphertext))

	mode.CryptBlocks(decrypted, ciphertext)

	return pkcs5Unpad(decrypted)
}

func pkcs5Pad(data []byte, blockSize int) []byte {

	padding := blockSize - len(data)%blockSize

	padtext := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(data, padtext...)
}

func pkcs5Unpad(data []byte) ([]byte, error) {

	length := len(data)

	if length == 0 {

		return nil, errors.New("invalid padding size")
	}

	padlen := int(data[length-1])

	if padlen > length {

		return nil, errors.New("invalid padding")
	}

	return data[:length-padlen], nil
}
