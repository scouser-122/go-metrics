package handler

import (
	"net"
	"net/http"

	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/logger"
)

func TrustedMiddleware(h http.HandlerFunc, config *config.ServerConfig) http.HandlerFunc {
	var trustedSubnet *net.IPNet
	if config.TrustedSubnet != "" {
		_, subnet, err := net.ParseCIDR(config.TrustedSubnet)
		if err != nil {
			panic(err)
		}
		trustedSubnet = subnet
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if trustedSubnet != nil && !isTrustedIP(r, trustedSubnet) {
			http.Error(w, "Forbidden IP", http.StatusForbidden)
			return
		}
		h(w, r)
	}
}

func isTrustedIP(r *http.Request, trustedSubnet *net.IPNet) bool {
	realIP := r.Header.Get("X-Real-IP")
	if realIP == "" {
		logger.Sugar.Errorf("X-Real-IP absent")
		return false
	}
	ip := net.ParseIP(realIP)
	if ip == nil {
		logger.Sugar.Errorf("incorrect X-Real-IP: %s", realIP)
		return false
	}
	if !trustedSubnet.Contains(ip) {
		logger.Sugar.Errorf("untrusted IP: %s", realIP)
		return false
	}
	return true
}
