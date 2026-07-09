package middleware

import (
	"errors"
	"fmt"
	"go-app/global"
	"go-app/pkg/response"
	"go-app/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthNMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		accCache, ok := global.Cache.Get("access_option")
		if !ok {
			response.ErrorResponse(
				c,
				response.ErrCodeCache,
			)
			c.Abort()
			return
		}
		accTokenOption := accCache.(utils.SecretKey)

		var tokenStr string
		token := c.GetHeader("Authorization")

		if token != "" {
			parts := strings.Split(strings.TrimSpace(token), " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenStr = parts[1]
			}
		} else {
			// Thử lấy từ query parameter cho trường hợp WebSocket
			tokenStr = c.Query("token")
		}

		if tokenStr == "" {
			response.ErrorResponse(
				c,
				response.ErrCodeTokenNotFound,
			)
			c.Abort()
			return
		}

		claims, err := utils.VerifyJWT(tokenStr, accTokenOption)
		if err != nil {
			fmt.Println("err: ", err)
			if errors.Is(err, jwt.ErrTokenExpired) {
				response.ErrorResponse(
					c,
					response.ErrCodeTokenExpired,
				)
			} else {
				response.ErrorResponse(
					c,
					response.ErrCodeTokenInvalid,
				)
			}
			c.Abort()
			return
		}
		// set thông tin user vào trong context để có thể sử dụng ở các handler khác
		c.Set("user-id", claims.UserInfo.UserId)
		c.Set("role", claims.UserInfo.Role)
		c.Set("email", claims.UserInfo.Email)
		c.Next()
	}
}
