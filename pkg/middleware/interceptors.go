package middleware //interceptors

import (
	"strings"

	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func getMethod(info *grpc.UnaryServerInfo) string {
	splits := strings.Split(info.FullMethod, "/")
	return splits[len(splits)-1]
}

func GetInterceptors(logger *zap.SugaredLogger) grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(
		kitgrpc.Interceptor,
		recoveryInterceptor(logger),
		loggingInterceptor(logger),
	)
}
