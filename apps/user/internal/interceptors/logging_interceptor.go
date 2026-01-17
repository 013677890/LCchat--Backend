package interceptors

import (
	"context"
	"time"

	"ChatServer/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LoggingUnaryInterceptor 记录基础日志（方法、耗时、错误码、trace_id）。
func LoggingUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()
		resp, err = handler(ctx, req)
		code := status.Code(err)

		if err != nil {
			logger.Warn(ctx, "grpc unary request",
				logger.String("method", info.FullMethod),
				logger.Duration("cost", time.Since(start)),
				logger.String("code", code.String()),
				logger.ErrorField("error", err),
			)
		} else {
			logger.Info(ctx, "grpc unary request",
				logger.String("method", info.FullMethod),
				logger.Duration("cost", time.Since(start)),
				logger.String("code", code.String()),
			)
		}

		return resp, err
	}
}
