package pb

import (
	userpb "ChatServer/apps/user/pb"
	"context"
	"time"

	"ChatServer/apps/gateway/internal/middleware"
	"ChatServer/pkg/logger"

	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// userServiceClientImpl 用户服务 gRPC 客户端实现
type userServiceClientImpl struct {
	client  userpb.UserServiceClient
	breaker *gobreaker.CircuitBreaker
}

// NewUserServiceClient 创建用户服务 gRPC 客户端实例
// conn: gRPC 连接
// breaker: 熔断器实例
func NewUserServiceClient(conn *grpc.ClientConn, breaker *gobreaker.CircuitBreaker) UserServiceClient {
	return &userServiceClientImpl{
		client:  userpb.NewUserServiceClient(conn),
		breaker: breaker,
	}
}

// Login 登录方法实现
// ctx: 上下文
// req: 登录请求
// 返回: 登录响应和错误
func (c *userServiceClientImpl) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	// 记录 gRPC 调用开始时间，用于计算耗时
	start := time.Now()

	// 使用熔断器包装 gRPC 调用
	var resp *userpb.LoginResponse
	var err error

	_, breakerErr := c.breaker.Execute(func() (interface{}, error) {
		resp, err = c.client.Login(ctx, req)
		return resp, err
	})

	// 如果熔断器返回错误（如熔断器开启），使用熔断器错误
	if breakerErr != nil {
		err = breakerErr
	}

	// 计算耗时并记录到 Prometheus 指标
	duration := time.Since(start).Seconds()
	middleware.RecordGRPCRequest("user.UserService", "Login", duration, err)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// ==================== gRPC 连接和熔断器初始化工具函数 ====================

// gRPC 服务配置，定义重试策略
const retryPolicy = `{
	"methodConfig": [{
		"name": [{"service": "user.UserService"}],
		"waitForReady": true,
		"timeout": "2s",
		"retryPolicy": {
			"maxAttempts": 5,
			"initialBackoff": "0.1s",
			"maxBackoff": "1s",
			"backoffMultiplier": 2,
			"retryableStatusCodes": ["UNAVAILABLE", "DEADLINE_EXCEEDED", "UNKNOWN"]
		}
	}]
}`

// CreateUserServiceConnection 创建用户服务 gRPC 连接
// addr: 用户服务地址，格式为 "host:port"
// breaker: 熔断器实例
// 返回: gRPC 连接和错误
func CreateUserServiceConnection(addr string, breaker *gobreaker.CircuitBreaker) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(retryPolicy), // 应用重试策略
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(4*1024*1024), // 4MB接收大小
		),
		// 注入熔断拦截器
		grpc.WithChainUnaryInterceptor(
			middleware.CircuitBreakerInterceptor(breaker),
		),
	)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// CreateCircuitBreaker 创建熔断器实例
// name: 熔断器名称
// 返回: 熔断器实例
func CreateCircuitBreaker(name string) *gobreaker.CircuitBreaker {
	return gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        name,
		MaxRequests: 3,                // 半开状态下最多允许 3 个请求尝试
		Interval:    15 * time.Second, // 清除计数的时间间隔
		Timeout:     45 * time.Second, // 熔断器开启后多久尝试进入半开状态
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// 失败率超过 50% 且连续失败次数超过 5 次时触发熔断
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 5 && failureRatio >= 0.5
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			logger.Info(context.Background(), "熔断器状态变化",
				logger.String("name", name),
				logger.String("from", from.String()),
				logger.String("to", to.String()),
			)
		},
	})
}
