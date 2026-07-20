package middleware

import (
	"time"

	"go-app/global"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CorsMiddleware() gin.HandlerFunc {
	// sử dụng thư viện gin-contrib/cors
	return cors.New(cors.Config{
		// chỉ cho phép các origin sau truy cập
		AllowOrigins: global.Config.Cors.AllowOrigins,
		// chỉ cho phép các method
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		// chỉ cho phép các header sau
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
		},
		// cho phép các header sau được trả về
		ExposeHeaders: []string{
			"Content-Length",
		},
		// cho phép trình duyệt truy cập vào
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
