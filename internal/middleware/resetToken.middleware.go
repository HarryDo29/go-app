package middleware

import (
	"fmt"
	"go-app/global"
	"go-app/pkg/response"
	"go-app/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func ResetTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		resetCache, ok := global.Cache.Get("reset_option")
		if !ok {
			response.ErrorResponse(
				c,
				response.ErrCodeTokenNotFound,
			)
			c.Abort()
			return
		}

		resetTokenOption := resetCache.(utils.SecretKey)
		token := c.GetHeader("Authorization")
		if token == "" {
			response.ErrorResponse(
				c,
				response.ErrCodeTokenNotFound,
			)
			c.Abort()
			return
		}

		parts := strings.Split(strings.TrimSpace(token), " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.ErrorResponse(
				c,
				response.ErrCodeTokenInvalid,
			)
			c.Abort()
			return
		}

		claims, err := utils.VerifyJWT(parts[1], resetTokenOption)
		if err != nil {
			fmt.Println("err: ", err)
			response.ErrorResponse(
				c,
				response.ErrCodeTokenInvalid,
			)
			c.Abort()
			return
		}

		// set thông tin user vào trong context để có thể sử dụng ở các handler khác
		c.Set("user-id", claims.UserInfo.UserId)
		c.Set("email", claims.UserInfo.Email)

		c.Next()
	}
}
