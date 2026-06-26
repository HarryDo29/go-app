package middleware

import (
	"go-app/global"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Cho request đi tiếp tới middleware/handler tiếp theo
		c.Next()

		// Sau khi handler xử lý xong thì lấy status + thời gian xử lý
		statusCode := c.Writer.Status()
		latency := time.Since(startTime)

		// Ghi log request có cấu trúc bằng global.Logger
		global.Logger.Info("[HTTP REQUEST]",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("ip", clientIP),
			zap.String("userAgent", userAgent),
		)
	}
}
