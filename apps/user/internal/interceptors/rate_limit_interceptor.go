package interceptors

import (
	"context"
	"sync"

	"ChatServer/pkg/logger"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RateLimiterConfig 限流器配置
type RateLimiterConfig struct {
	// RequestsPerSecond 每秒允许的请求数
	RequestsPerSecond float64
	// Burst 允许的突发请求数（令牌桶容量）
	Burst int
}

// DefaultRateLimiterConfig 默认限流配置
var DefaultRateLimiterConfig = RateLimiterConfig{
	RequestsPerSecond: 1500, // 每秒 1500 个请求
	Burst:             2500, // 允许 2500 个突发请求
}

// rateLimiter 基于令牌桶算法的全局限流器
type rateLimiter struct {
	limiter *rate.Limiter
	config  RateLimiterConfig
}

// globalRateLimiter 全局限流器实例（单例模式）
var (
	globalRateLimiter *rateLimiter
	once              sync.Once
)

// getGlobalRateLimiter 获取全局限流器实例
func getGlobalRateLimiter(config RateLimiterConfig) *rateLimiter {
	once.Do(func() {
		globalRateLimiter = &rateLimiter{
			limiter: rate.NewLimiter(rate.Limit(config.RequestsPerSecond), config.Burst),
			config:  config,
		}
	})
	return globalRateLimiter
}

// RateLimitUnaryInterceptor 创建限流拦截器
// 使用令牌桶算法实现全局限流，防止服务被突发流量击垮
func RateLimitUnaryInterceptor(config ...RateLimiterConfig) grpc.UnaryServerInterceptor {
	cfg := DefaultRateLimiterConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	limiter := getGlobalRateLimiter(cfg)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// 尝试从令牌桶获取令牌
		if !limiter.limiter.Allow() {
			logger.Warn(ctx, "请求被限流拦截",
				logger.String("method", info.FullMethod),
				logger.Float64("limit_rate", cfg.RequestsPerSecond),
				logger.Int("burst", cfg.Burst),
			)
			// 返回资源耗尽错误
			return nil, status.Error(codes.ResourceExhausted, "服务繁忙，请稍后重试")
		}

		// 获取令牌成功，执行业务逻辑
		return handler(ctx, req)
	}
}
