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

		resp, err := handler(ctx, req)
		logger.Infow("spawnerservice", "method", "CreateCluster", "request", request, "response", resp, "error", err, "took", time.Since(begin))
		return nil, nil
	}
}
