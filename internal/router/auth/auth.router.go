package auth

import (
	"go-app/internal/middleware"
	"go-app/internal/wire"

	"github.com/gin-gonic/gin"
)

type AuthRouter struct{}

func (ur *AuthRouter) InitAuthRouter(router *gin.RouterGroup) {
	// wired (dependency injection DI)
	authController, _ := wire.InitAuthRouterHandler()

	// === PUBLIC routes (không cần JWT) ===
	authRouterPublic := router.Group("/auth")
	{
		authRouterPublic.POST("/register", authController.Register)
		authRouterPublic.POST("/login", authController.Login)
		authRouterPublic.POST("/refresh", authController.RefreshToken)
		// Forget-password flow (public)
		authRouterPublic.POST("/forget", authController.ForgetPassword)
		authRouterPublic.POST("/forget/verify", authController.VerifyForgetPassword)
		// Reset-password: dùng reset_token từ Authorization header (không phải JWT)
		authRouterPublic.POST("/reset",
			middleware.ResetTokenMiddleware(),
			authController.ResetPassword,
		)
	}

	// === PRIVATE routes (cần JWT) ===
	authRouterPrivate := router.Group("/auth")
	authRouterPrivate.Use(middleware.AuthNMiddleware())
	{
		authRouterPrivate.POST("/logout", authController.Logout)
		authRouterPrivate.POST("/change-password", authController.ChangePassword)
		authRouterPrivate.POST("/otp", authController.SendOtp)            // yêu cầu JWT
		authRouterPrivate.POST("/otp/verify", authController.ValidateOtp) // yêu cầu JWT
	}
}
