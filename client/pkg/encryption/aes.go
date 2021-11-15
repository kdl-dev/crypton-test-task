package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"errors"
	"golang.org/x/crypto/pbkdf2"
	"os"
)

func GenerateKey(password []byte) []byte {
	return pbkdf2.Key(password, []byte(os.Getenv("SALT")), 4096, 32, sha1.New)
}

func Encode(data []byte, password string, mode string) ([]byte, error) {
	key := GenerateKey([]byte(password))
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	switch mode {
	case Cbc:
		if encryptedData, err := cbcEncode(data, block); err != nil {
			return nil, err
		} else { return encryptedData, nil }
	default:
		return nil, errors.New("unsupported or unknown encryption mode")
	}
}

func Decode(data []byte, password string, mode string) ([]byte, error) {
	key := GenerateKey([]byte(password))
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	switch mode {
	case Cbc:
		if decryptedData, err := cbcDecode(data, block); err != nil {
			return nil, err
		} else { return decryptedData, nil }
	default:
		return nil, errors.New("unsupported or unknown encryption mode")
	}
}

func cbcEncode(data []byte, block cipher.Block) ([]byte, error) {
	data = Pkcs5Padding(data, block.BlockSize())
	encryptedData := make([]byte, len(data))
	iv := []byte(os.Getenv("IV"))

	if len(iv) != 16 {
		return nil, errors.New("the AES algorithm requires that the IV size must be 16 bytes")
	}

	blockMode := cipher.NewCBCEncrypter(block, iv)
	blockMode.CryptBlocks(encryptedData, data)
	return encryptedData, nil
}

func cbcDecode(encryptedData []byte, block cipher.Block) ([]byte, error) {
	decryptedData := make([]byte, len(encryptedData))
	iv := []byte(os.Getenv("IV"))

	if len(iv) != 16 {
		return nil, errors.New("the AES algorithm requires that the IV size must be 16 bytes")
	}

	blockMode := cipher.NewCBCDecrypter(block, iv)
	blockMode.CryptBlocks(decryptedData, encryptedData)
	decryptedData = pkcs5Trimming(decryptedData)

	return decryptedData, nil
}

func Pkcs5Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	paddingText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, paddingText...)
}

func pkcs5Trimming(encryptedData []byte) []byte {
	padding := encryptedData[len(encryptedData)-1]
	return encryptedData[:len(encryptedData)-int(padding)]
}

func EncodeChunkSize(dataChuckSize int, password string) int {
	key := GenerateKey([]byte(password))
	block, err := aes.NewCipher(key)
	if err != nil {
		return -1
	}

	blockSize := block.BlockSize()
	return dataChuckSize + blockSize - dataChuckSize % blockSize
}
