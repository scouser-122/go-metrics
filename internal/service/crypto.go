package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/logger"
)

// CryptoService provides cryptographic operations for request/response verification.
type CryptoService struct {
	ServerConfig *config.ServerConfig
	privateKey   *rsa.PrivateKey
}

// KeyPresent checks if an HMAC key is configured.
func (service *CryptoService) KeyPresent() bool {
	return service.ServerConfig.HMACKey != ""
}

// CalculateHash calculates the HMAC-SHA256 hash of the provided data.
func (service *CryptoService) CalculateHash(src []byte) string {
	hmac := hmac.New(sha256.New, []byte(service.ServerConfig.HMACKey))
	hmac.Write(src)
	return hex.EncodeToString(hmac.Sum(nil))
}

// LoadPrivateKeyIfExists loads private key if it exists in FS,
// if loaded - will be used to encode requests to server
func (service *CryptoService) LoadPrivateKeyIfExists() {
	if service.ServerConfig.CryptoKey == "" {
		logger.Sugar.Info("crypto key path not specified")
		return
	}
	lastSlash := strings.LastIndex(service.ServerConfig.CryptoKey, "/")
	if lastSlash < 0 {
		logger.Sugar.Errorf("crypto key path not correct: %s", service.ServerConfig.CryptoKey)
		return
	}
	dirPath := service.ServerConfig.CryptoKey[:lastSlash]

	root, err := os.OpenRoot(dirPath)
	if err != nil {
		logger.Sugar.Error(err)
		return
	}
	defer root.Close()

	fileName := service.ServerConfig.CryptoKey[lastSlash+1:]
	file, err := root.Open(fileName)
	if err != nil {
		logger.Sugar.Error(err)
		return
	}
	defer file.Close()

	privateKeyBytes, err := io.ReadAll(file)
	if err != nil {
		logger.Sugar.Error(err)
		return
	}

	block, _ := pem.Decode(privateKeyBytes)
	if block == nil {
		logger.Sugar.Errorf("failed to decode private key PEM block")
		return
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		logger.Sugar.Errorf("failed to parse private key: %w", err)
		return
	}

	// Type assert to RSA public key
	rsaPrivate, ok := key.(*rsa.PrivateKey)
	if !ok {
		logger.Sugar.Errorf("not an RSA public key")
		return
	}

	service.privateKey = rsaPrivate
	logger.Sugar.Info("successfully loaded private key")
}

// DecryptRequestBody decrypts request body if private key was specified in config,
// if key not specified - returns same body bytes which were passed
func (service *CryptoService) DecryptRequestBody(encryptedBody []byte) ([]byte, error) {
	if service.privateKey == nil {
		return encryptedBody, nil
	}

	chunkSize := service.privateKey.Size()

	// If the encrypted message is the same size as a single chunk, decrypt directly
	if len(encryptedBody) <= chunkSize {
		return rsa.DecryptOAEP(sha256.New(), rand.Reader, service.privateKey, encryptedBody, nil)
	}

	// Split the encrypted message into chunks
	var decryptedData []byte
	for start := 0; start < len(encryptedBody); start += chunkSize {
		end := start + chunkSize
		if end > len(encryptedBody) {
			// If the last chunk is smaller than chunkSize, it might be a direct encryption
			// or we have incomplete data
			decryptedChunk, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, service.privateKey, encryptedBody[start:], nil)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt final chunk: %w", err)
			}
			decryptedData = append(decryptedData, decryptedChunk...)
			break
		}

		chunk := encryptedBody[start:end]
		decryptedChunk, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, service.privateKey, chunk, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt chunk: %w", err)
		}

		decryptedData = append(decryptedData, decryptedChunk...)
	}

	return decryptedData, nil
}
