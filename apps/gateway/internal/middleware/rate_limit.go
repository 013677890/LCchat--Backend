package middleware

import (
	"ChatServer/pkg/logger"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// UserRateLimiter 用户级别的限流器
// 为每个用户维护独立的令牌桶
type UserRateLimiter struct {
	limiters map[string]*rate.Limiter // key: user_uuid, value: 令牌桶
	mu       *sync.RWMutex
	r        rate.Limit // 每秒产生的令牌数
	b        int        // 令牌桶容量
}

// NewUserRateLimiter 创建用户级别限流器
// requestsPerSecond: 每秒允许的请求数（令牌产生速率）
// burst: 令牌桶容量（允许的突发请求数）
func NewUserRateLimiter(requestsPerSecond float64, burst int) *UserRateLimiter {
	return &UserRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		mu:       &sync.RWMutex{},
		r:        rate.Limit(requestsPerSecond),
		b:        burst,
	}
}

// GetLimiter 获取指定用户的限流器
// 如果用户的限流器不存在，则创建一个新的
func (u *UserRateLimiter) GetLimiter(userUUID string) *rate.Limiter {
	u.mu.Lock()
	defer u.mu.Unlock()

	limiter, exists := u.limiters[userUUID]
	if !exists {
		// 为新用户创建令牌桶
		limiter = rate.NewLimiter(u.r, u.b)
		u.limiters[userUUID] = limiter
	}

	return limiter
}

// CleanupInactiveLimiters 清理长时间未使用的限流器
// 定期调用此方法可以释放内存
func (u *UserRateLimiter) CleanupInactiveLimiters(inactiveDuration time.Duration) {
	u.mu.Lock()
	defer u.mu.Unlock()

	for userUUID, limiter := range u.limiters {
		// 检查令牌桶是否长时间未使用
		// 如果令牌桶已满，说明很久没有请求了
		if limiter.Tokens() >= float64(u.b) {
			// 简单策略：删除令牌桶已满的用户
			// 更精确的做法需要记录最后使用时间
			delete(u.limiters, userUUID)
		}
	}
}

// GetLimiterCount 获取当前限流器数量（用于监控）
func (u *UserRateLimiter) GetLimiterCount() int {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return len(u.limiters)
}

// 全局用户限流器实例
var globalUserLimiter *UserRateLimiter

// InitUserRateLimiter 初始化全局用户限流器
func InitUserRateLimiter(requestsPerSecond float64, burst int) {
	globalUserLimiter = NewUserRateLimiter(requestsPerSecond, burst)

	// 启动定期清理协程（每小时清理一次）
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			if globalUserLimiter != nil {
				globalUserLimiter.CleanupInactiveLimiters(30 * time.Minute)
			}
		}
	}()
}

// UserRateLimitMiddleware 用户级别限流中间件
// 必须在 JWT 认证中间件之后使用，因为需要从 context 中获取 user_uuid
func UserRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 context 中获取用户 UUID（由 JWT 中间件设置）
		userUUID, exists := GetUserUUID(c)
		if !exists || userUUID == "" {
			// 如果没有用户信息，说明是公开接口或者认证失败
			// 这种情况应该已经被前面的中间件拦截了
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证，无法进行限流检查",
			})
			c.Abort()
			return
		}

		// 获取该用户的限流器
		limiter := globalUserLimiter.GetLimiter(userUUID)

		// 尝试获取令牌
		if !limiter.Allow() {
			// 没有可用令牌，请求被限流
			logger.Warn(c.Request.Context(), "用户请求被限流",
				logger.String("user_uuid", userUUID),
				logger.String("path", c.Request.URL.Path),
				logger.String("method", c.Request.Method),
			)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}

		// 通过限流检查，继续处理请求
		c.Next()
	}
}

// UserRateLimitMiddlewareWithConfig 可配置的用户限流中间件
// 允许为不同的路由组设置不同的限流参数
func UserRateLimitMiddlewareWithConfig(requestsPerSecond float64, burst int) gin.HandlerFunc {
	// 创建独立的限流器实例
	limiter := NewUserRateLimiter(requestsPerSecond, burst)

	return func(c *gin.Context) {
		userUUID, exists := GetUserUUID(c)
		if !exists || userUUID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证，无法进行限流检查",
			})
			c.Abort()
			return
		}

		userLimiter := limiter.GetLimiter(userUUID)

		if !userLimiter.Allow() {
			logger.Warn(c.Request.Context(), "用户请求被限流",
				logger.String("user_uuid", userUUID),
				logger.String("path", c.Request.URL.Path),
				logger.String("method", c.Request.Method),
			)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
