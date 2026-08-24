package encryptor

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

var (
	ErrInvalidKeySize     = errors.New("panjang kunci enkripsi harus tepat 32 byte untuk AES-256")
	ErrCiphertextTooShort = errors.New("teks ciphertext terlalu pendek")
	ErrDecryptionFailed   = errors.New("gagal mendekripsi data atau otentikasi tag gagal")
)

func deriveKey(key []byte) []byte {
	if len(key) == 32 {
		return key
	}
	if len(key) == 64 {
		if decoded, err := hex.DecodeString(string(key)); err == nil && len(decoded) == 32 {
			return decoded
		}
	}
	h := sha256.Sum256(key)
	return h[:]
}

// Encrypt mengenkripsi teks mentah menggunakan algoritma AES-256-GCM dengan nonce acak 12-byte.
// Parameter plainText merupakan teks rahasia yang akan dienkripsi.
// Parameter key merupakan byte slice kunci rahasia sepanjang 32-byte (256-bit) atau 64-char hex string.
// Mengembalikan string ciphertext berformat base64 (nonce + tag + ciphertext) atau error jika enkripsi gagal.
func Encrypt(plainText string, key []byte) (string, error) {
	aesKey := deriveKey(key)

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("gagal membuat blok chipper: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("gagal membuat mode GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("gagal menghasilkan nonce acak: %w", err)
	}

	cipherBytes := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(cipherBytes), nil
}

// Decrypt mendekripsi string base64 ciphertext yang dihasilkan oleh fungsi Encrypt.
// Parameter cipherText merupakan string terenkripsi berformat base64.
// Parameter key merupakan byte slice kunci rahasia.
// Mengembalikan teks asli (plainText) atau error jika integritas data tidak valid atau kunci salah.
func Decrypt(cipherText string, key []byte) (string, error) {
	aesKey := deriveKey(key)

	data, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", fmt.Errorf("gagal decode base64 ciphertext: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("gagal membuat blok cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("gagal membuat mode GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrCiphertextTooShort
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plainBytes, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plainBytes), nil
}
