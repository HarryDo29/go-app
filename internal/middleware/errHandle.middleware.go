package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorHandlerMiddleware: middleware để xử lý lỗi và trả về response
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			// recover: phục hồi lại chương trình khi bị panic
			if err := recover(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "Something went wrong",
				})
				c.Abort()
				return
			}
		}()

		// next: gọi đến handler tiếp theo
		c.Next()

		// nếu có lỗi trong context
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			c.JSON(c.Writer.Status(), gin.H{
				"success": false,
				"code":    "REQUEST_ERROR",
				"message": err.Error(),
			})
			return
		}
	}
}
