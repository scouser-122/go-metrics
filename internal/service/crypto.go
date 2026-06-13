package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/scouser-122/go-metrics/internal/config"
)

// CryptoService provides cryptographic operations for request/response verification.
type CryptoService struct {
	ServerConfig *config.ServerConfig
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
