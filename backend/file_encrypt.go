// File encryption and decryption

package main

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io/ioutil"
)

type FileEncryptionMethod uint16

const (
	AES256_ZIP  FileEncryptionMethod = 1 // Compress data, then encrypt it
	AES256_FLAT FileEncryptionMethod = 2 // Just encrypt the data
)

// Encrypts file contents
// data - File data
// method - algorithm to use
// key - Encryption key
// Returns the cipher text
func encryptFileContents(data []byte, method FileEncryptionMethod, key []byte) ([]byte, error) {
	if len(data) == 0 {
		return make([]byte, 0), nil
	}

	result := make([]byte, 2)

	binary.BigEndian.PutUint16(result, uint16(method)) // Include method

	if method == AES256_ZIP {
		// Compress the data
		var b bytes.Buffer
		w := zlib.NewWriter(&b)
		w.Write(data)
		w.Close()
		finalData := b.Bytes()

		// Include pre-encryption size to the header
		header := make([]byte, 20)
		binary.BigEndian.PutUint32(header[:4], uint32(len(finalData)))

		// Pad data
		finalData = PKCS5Padding(finalData, 16)

		// Generate IV
		iv := make([]byte, 16)
		rand.Read(iv)

		// Include IV into the header
		copy(header[4:20], iv)

		// Encrypt
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
		ciphertext := make([]byte, len(finalData))
		mode := cipher.NewCBCEncrypter(block, iv)
		mode.CryptBlocks(ciphertext, finalData)

		// Include in result

		result = append(result, header...)
		result = append(result, ciphertext...)
	} else if method == AES256_FLAT {
		// Include pre-encryption size to the header
		header := make([]byte, 20)
		binary.BigEndian.PutUint32(header[:4], uint32(len(data)))

		// Pad data
		finalData := PKCS5Padding(data, 16)

		// Generate IV
		iv := make([]byte, 16)
		rand.Read(iv)

		// Include IV into the header
		copy(header[4:20], iv)

		// Encrypt
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
		ciphertext := make([]byte, len(finalData))
		mode := cipher.NewCBCEncrypter(block, iv)
		mode.CryptBlocks(ciphertext, finalData)

		// Include in result

		result = append(result, header...)
		result = append(result, ciphertext...)

	} else {
		return nil, errors.New("Invalid method")
	}

	return result, nil
}

// Decripts file contents
// data - Cipher text
// key - Decryption key
// Returns the original file data
func decryptFileContents(data []byte, key []byte) ([]byte, error) {
	if len(data) < 2 {
		if len(data) == 0 {
			return make([]byte, 0), nil
		} else {
			return nil, errors.New("Invalid data provided")
		}
	}

	method := FileEncryptionMethod(binary.BigEndian.Uint16(data[:2]))

	if method == AES256_ZIP {
		if len(data) < 23 {
			return nil, errors.New("Invalid data provided")
		}

		// Read params
		preEncDataLength := int(binary.BigEndian.Uint32(data[2:6]))
		iv := data[6:22]
		ciphertext := data[22:]

		if preEncDataLength < 0 || preEncDataLength > len(ciphertext) {
			return nil, errors.New("Invalid method")
		}

		// Decrypt
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
		mode := cipher.NewCBCDecrypter(block, iv)
		plaintext := make([]byte, len(ciphertext))
		mode.CryptBlocks(plaintext, ciphertext)

		// Remove padding
		plaintext = plaintext[:preEncDataLength]

		// Decompress the data
		bSource := bytes.NewReader(plaintext)
		r, err := zlib.NewReader(bSource)
		if err != nil {
			return nil, err
		}
		result, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		r.Close()

		return result, nil
	} else if method == AES256_FLAT {
		if len(data) < 23 {
			return nil, errors.New("Invalid data provided")
		}

		// Read params
		preEncDataLength := int(binary.BigEndian.Uint32(data[2:6]))
		iv := data[6:22]
		ciphertext := data[22:]

		if preEncDataLength < 0 || preEncDataLength > len(ciphertext) {
			return nil, errors.New("Invalid method")
		}

		// Decrypt
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
		mode := cipher.NewCBCDecrypter(block, iv)
		plaintext := make([]byte, len(ciphertext))
		mode.CryptBlocks(plaintext, ciphertext)

		// Remove padding
		plaintext = plaintext[:preEncDataLength]

		return plaintext, nil
	} else {
		return nil, errors.New("Invalid method")
	}
}

// Add padding to the data, so a block cipher can encrypt it
// ciphertext - data
// blockSize - Size of the blocks to encrypt
func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := (blockSize - len(ciphertext)%blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
