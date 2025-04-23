package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/axelburling/dfs/internal/log"
	"go.uber.org/zap"
)

type Crypto struct {
	key []byte
	log *log.Logger
}

func NewCrypto(key []byte, log *log.Logger) (*Crypto, error) {
	if len(key) != 32 {
		log.Error("invalid AES key length", zap.Int("length", len(key)))
		return nil, fmt.Errorf("aes key needs to be 32 bytes to use the AES-256")
	}

	log.Debug("initialized Crypto")
	return &Crypto{
		key: key,
		log: log,
	}, nil
}

func (c *Crypto) encrypt(data []byte) (io.Reader, error) {
	c.log.Debug("encrypting data", zap.Int("size", len(data)))
	ci, err := aes.NewCipher(c.key)

	if err != nil {
		c.log.Error("failed to create AES cipher", zap.Error(err))
		return nil, err
	}

	gcm, err := cipher.NewGCM(ci)

	if err != nil {
		c.log.Error("failed to create GCM mode", zap.Error(err))
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
    _, err = rand.Read(nonce)
    if err != nil {
		c.log.Error("failed to generate nonce", zap.Error(err))
        return nil, err
    }

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	c.log.Debug("encryption successful", zap.Int("encrypted_size", len(ciphertext)))

	return bytes.NewReader(ciphertext), nil
}

func (c *Crypto) Encrypt(data []byte) (io.Reader, error) {
	return c.encrypt(data)
}

func (c *Crypto) EncryptReader(reader io.Reader) (io.Reader, error) {
	data, err := io.ReadAll(reader)

	if err != nil {
		c.log.Error("failed to read from input reader", zap.Error(err))
		return nil, err
	}

	return c.encrypt(data)
}

func (c *Crypto) Decrypt(reader io.ReadCloser) (io.Reader, error) {
	data, err := io.ReadAll(reader)

	if err != nil {
		c.log.Error("failed to read encrypted data", zap.Error(err))
		return nil, err
	}

	aes, err := aes.NewCipher(c.key)
    if err != nil {
		c.log.Error("failed to create AES cipher for decryption", zap.Error(err))
        return nil, err
    }

    gcm, err := cipher.NewGCM(aes)
    if err != nil {
		c.log.Error("failed to create GCM mode for decryption", zap.Error(err))
        return nil, err
    }

    nonceSize := gcm.NonceSize()

	if len(data) < nonceSize {
		c.log.Error("invalid encrypted data", zap.Int("size", len(data)), zap.Int("nonce_size", nonceSize))
		return nil, fmt.Errorf("invalid encrypted data")
	}

    nonce, ciphertext := data[:nonceSize], data[nonceSize:]

    plaintext, err := gcm.Open(nil, []byte(nonce), []byte(ciphertext), nil)
    if err != nil {
        return nil, err
    }

    return bytes.NewReader(plaintext), nil
}