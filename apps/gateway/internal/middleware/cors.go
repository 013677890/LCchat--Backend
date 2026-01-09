package middleware

import (
	"github.com/gin-gonic/gin"
)

// CorsMiddleware 跨域中间件
func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 定义允许的白名单
		//allowedOrigins := map[string]bool{
		//	"http://localhost:8080": true, // Web 开发
		//	"https://my-web.com":    true, // Web 生产
		//	"app://my-app":          true, // Electron 自定义协议
		//	"null":                  true, // 某些 Electron 环境下 Origin 可能是 null (慎用)
		//}

		//测试环境 全部允许（带凭据）
		c.Header("Access-Control-Allow-Origin", origin) // 返回请求的具体 Origin，不能是 *
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, x-requested-with")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Vary", "Origin") // 重要：告诉浏览器 Origin 值会变化

		// 处理 OPTIONS 预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
