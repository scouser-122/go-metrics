package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/scouser-122/go-metrics/internal/config"
)

type CryptoService struct {
	ServerConfig *config.ServerConfig
}

func (service *CryptoService) KeyPresent() bool {
	return service.ServerConfig.HMACKey != ""
}

func (service *CryptoService) CalculateHash(src []byte) string {
	hmac := hmac.New(sha256.New, []byte(service.ServerConfig.HMACKey))
	hmac.Write(src)
	return hex.EncodeToString(hmac.Sum(nil))
}
