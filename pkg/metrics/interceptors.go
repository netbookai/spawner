package metrics

import (
	"context"
	"strings"

	"google.golang.org/grpc"
)

func getMethod(info *grpc.UnaryServerInfo) string {
	splits := strings.Split(info.FullMethod, "/")
	return splits[len(splits)-1]
}

//RPCInstrumentation request instrumentation interceptor
func RPCInstrumentation() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		IncRequest(getMethod(info))
		return handler(ctx, req)
	}
}
