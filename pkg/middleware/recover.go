package middleware

// Interceptors

import (
	"context"
	"runtime/debug"

	"github.com/gogo/status"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func recoveryInterceptor(logger *zap.SugaredLogger) grpc.UnaryServerInterceptor {

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		panicked := true

		defer func() {
			if r := recover(); r != nil || panicked {
				//log error details and stack trace
				methodName := getMethod(info)
				logger.Errorf("failed to handle the request [PANIC]", "method", methodName, "error", err, "stacktrace", string(debug.Stack()))
				err = status.Errorf(codes.Internal, "%v in call to method '%s'", r, methodName)
			}
		}()

		resp, err := handler(ctx, req)
		panicked = false
		return resp, err
	}
}
