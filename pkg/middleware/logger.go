package middleware

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func loggingInterceptor(logger *zap.SugaredLogger) grpc.UnaryServerInterceptor {

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		// log request and response data

		begin := time.Now()
		request := fmt.Sprintf("%+v", req)
		method := getMethod(info)

		logger.Debugw("got request in spawnerservice", "method", method, "request", request)
		resp, err := handler(ctx, req)
		logger.Infow("spawnerservice", "method", method, "request", request, "response", resp, "error", err, "took", time.Since(begin))
		return resp, err
	}
}
