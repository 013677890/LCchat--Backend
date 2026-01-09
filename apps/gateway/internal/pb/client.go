package pb

import (
	userpb "ChatServer/apps/user/pb"
	"context"
	"fmt"
	"time"

	"ChatServer/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	// userServiceClient 用户服务的gRPC客户端
	userServiceClient userpb.UserServiceClient
	// userServiceConn 用户服务的gRPC连接
	userServiceConn *grpc.ClientConn
)

// InitUserServiceClient 初始化用户服务gRPC客户端
// addr: 用户服务地址，格式为 "host:port"，例如 "localhost:9090"
func InitUserServiceClient(addr string) error {

	// 建立gRPC连接
	// 使用 insecure credentials（实际生产环境应该使用TLS）
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(4*1024*1024), // 4MB接收大小
		),
	)
	if err != nil {
		return err
	}

	userServiceConn = conn
	userServiceClient = userpb.NewUserServiceClient(conn)

	return nil
}

// CloseUserServiceClient 关闭用户服务gRPC客户端
func CloseUserServiceClient() error {
	if userServiceConn != nil {
		if err := userServiceConn.Close(); err != nil {
			return err
		}
	}

	return nil
}

// GetUserServiceClient 获取用户服务gRPC客户端
func GetUserServiceClient() userpb.UserServiceClient {
	return userServiceClient
}

// Login 登录方法
// ctx: 上下文
// req: 登录请求
// 返回: 登录响应和错误
func Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	client := GetUserServiceClient()
	if client == nil {
		return nil, fmt.Errorf("user service client not initialized")
	}

	resp, err := client.Login(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}


// loginWithRetry 带重试机制的登录方法
// ctx: 上下文
// req: 登录请求
// maxRetries: 最大重试次数
// 返回: 登录响应和错误
func LoginWithRetry(ctx context.Context, req *userpb.LoginRequest, maxRetries int) (*userpb.LoginResponse, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// 如果不是第一次尝试，添加延迟
		if attempt > 0 {
			// 指数退避: 100ms, 200ms, 400ms, ...
			backoff := time.Duration(100<<uint(attempt-1)) * time.Millisecond
			if backoff > time.Second {
				backoff = time.Second // 最大延迟1秒
			}

			select {
			case <-time.After(backoff):
				logger.Warn(ctx, "登录重试",
					logger.Int("attempt", attempt),
					logger.Duration("backoff", backoff),
				)
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		// 调用登录方法
		resp, err := Login(ctx, req)
		if err == nil {
			// 成功则直接返回
			return resp, nil
		}

		// 记录错误
		lastErr = err

		// 如果上下文已取消，直接返回错误
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}

	// 所有重试都失败
	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
