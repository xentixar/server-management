package helpers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// EncryptAESGCM seals plaintext with a 32-byte key. Ciphertext is base64(nonce|seal).
func EncryptAESGCM(key []byte, plaintext string) (string, error) {
	if len(key) != 32 {
		return "", errors.New("encryption key must be 32 bytes")
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
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// DecryptAESGCM opens ciphertext produced by EncryptAESGCM.
func DecryptAESGCM(key []byte, ciphertext string) (string, error) {
	if len(key) != 32 {
		return "", errors.New("encryption key must be 32 bytes")
	}
	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("invalid ciphertext: %w", err)
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
	if len(raw) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, sealed := raw[:nonceSize], raw[nonceSize:]
	plain, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// DecodeEncryptionKey parses a base64-encoded 32-byte AES key.
func DecodeEncryptionKey(encoded string) ([]byte, error) {
	if encoded == "" {
		return nil, errors.New("encryption key is not configured")
	}
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid encryption key encoding: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must decode to 32 bytes, got %d", len(key))
	}
	return key, nil
}
