package handler

import (
	"context"
	"net"
	"net/http"

	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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
		realIP := r.Header.Get("X-Real-IP")
		if trustedSubnet != nil && !isTrustedIP(realIP, trustedSubnet) {
			http.Error(w, "Forbidden IP", http.StatusForbidden)
			return
		}
		h(w, r)
	}
}

func TrustedGrpcMiddleware(config *config.ServerConfig) grpc.UnaryServerInterceptor {
	var trustedSubnet *net.IPNet
	if config.TrustedSubnet != "" {
		_, subnet, err := net.ParseCIDR(config.TrustedSubnet)
		if err != nil {
			panic(err)
		}
		trustedSubnet = subnet
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var ip string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			values := md.Get("X-Real-IP")
			if len(values) > 0 {
				ip = values[0]
			}
		}
		if len(ip) == 0 {
			return nil, status.Error(codes.PermissionDenied, "missing IP")
		}
		if trustedSubnet != nil && !isTrustedIP(ip, trustedSubnet) {
			return nil, status.Error(codes.PermissionDenied, "Forbidden IP")
		}
		return handler(ctx, req)
	}
}

func isTrustedIP(realIP string, trustedSubnet *net.IPNet) bool {
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
