package handler

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/scouser-122/go-metrics/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TrustedMiddleware(h http.HandlerFunc, trustedSubnet *net.IPNet) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if trustedSubnet != nil {
			realIP := r.Header.Get("X-Real-IP")
			if !isTrustedIP(fmt.Sprintf("HTTP request %s", r.URL.RawQuery), realIP, trustedSubnet) {
				http.Error(w, "Forbidden IP", http.StatusForbidden)
				return
			}
		}
		h(w, r)
	}
}

func TrustedGrpcMiddleware(trustedSubnet *net.IPNet) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if trustedSubnet != nil {
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
			if !isTrustedIP("gPRC request", ip, trustedSubnet) {
				return nil, status.Error(codes.PermissionDenied, "Forbidden IP")
			}
		}
		return handler(ctx, req)
	}
}

func isTrustedIP(request string, realIP string, trustedSubnet *net.IPNet) bool {
	if realIP == "" {
		logger.Sugar.Errorf("X-Real-IP absent in incoming %s request", request)
		return false
	}
	ip := net.ParseIP(realIP)
	if ip == nil {
		logger.Sugar.Errorf("incorrect X-Real-IP: %s in incoming %s request", realIP, request)
		return false
	}
	if !trustedSubnet.Contains(ip) {
		logger.Sugar.Errorf("untrusted IP: %s in incoming %s request", realIP, request)
		return false
	}
	return true
}
