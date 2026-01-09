package pb

import (
	userpb "ChatServer/apps/user/pb"
	"ChatServer/pkg/logger"
	"context"
	"fmt"
	"time"

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
	ctx := context.Background()

	logger.Info(ctx, "初始化用户服务 gRPC 客户端",
		logger.String("address", addr),
	)

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
		logger.Error(ctx, "创建 gRPC 连接失败",
			logger.ErrorField("error", err),
			logger.String("address", addr),
		)
		return fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	userServiceConn = conn
	userServiceClient = userpb.NewUserServiceClient(conn)

	logger.Info(ctx, "用户服务 gRPC 客户端初始化成功",
		logger.String("address", addr),
	)

	return nil
}

// CloseUserServiceClient 关闭用户服务gRPC客户端
func CloseUserServiceClient() error {
	ctx := context.Background()

	logger.Info(ctx, "关闭用户服务 gRPC 客户端")

	if userServiceConn != nil {
		if err := userServiceConn.Close(); err != nil {
			logger.Error(ctx, "关闭 gRPC 连接失败",
				logger.ErrorField("error", err),
			)
			return err
		}
		logger.Info(ctx, "用户服务 gRPC 客户端已关闭")
	}

	return nil
}

// GetUserServiceClient 获取用户服务gRPC客户端
func GetUserServiceClient() userpb.UserServiceClient {
	return userServiceClient
}

// Login 登录方法（不带重试）
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
		// 单次调用失败会被重试,记录警告日志
		logger.Warn(ctx, "调用用户服务登录失败",
			logger.ErrorField("error", err),
			logger.String("telephone", req.Telephone),
		)
		return nil, err
	}

	return resp, nil
}

// LoginWithRetry 带重试机制的登录方法
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

			logger.Debug(ctx, "延迟后重试登录请求",
				logger.Int("attempt", attempt+1),
				logger.Int("max_retries", maxRetries),
				logger.Duration("backoff", backoff),
			)

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		// 调用登录方法
		resp, err := Login(ctx, req)
		if err == nil {
			// 成功则直接返回
			if attempt > 0 {
				logger.Info(ctx, "登录重试成功",
					logger.Int("attempt", attempt+1),
					logger.String("telephone", req.Telephone),
				)
			}
			return resp, nil
		}

		// 记录错误
		lastErr = err
		logger.Warn(ctx, "登录尝试失败",
			logger.Int("attempt", attempt+1),
			logger.Int("max_retries", maxRetries),
			logger.ErrorField("error", err),
		)

		// 如果上下文已取消，直接返回错误
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}

	// 所有重试都失败
	logger.Error(ctx, "所有登录重试均失败",
		logger.Int("max_retries", maxRetries),
		logger.ErrorField("last_error", lastErr),
	)

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
