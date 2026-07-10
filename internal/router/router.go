package routers

import (
	"go-app/global"
	"go-app/internal/middleware"
	_ "go-app/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter() *gin.Engine {
	var r *gin.Engine

	if global.Config.Server.Mode == "dev" {
		gin.SetMode(gin.DebugMode)
		gin.ForceConsoleColor()
		r = gin.Default()
	} else {
		gin.SetMode(gin.ReleaseMode)
		r = gin.New()
	}

	// middlewares
	r.Use(middleware.LoggerMiddleware())                 // logging
	r.Use(middleware.CorsMiddleware())                   // cors
	r.Use(middleware.NewRateLimitMiddleware().Handler()) // limiter global

	// router group
	userRouter := RouterGroupApp.User
	authRouter := RouterGroupApp.Auth
	rfRouter := RouterGroupApp.Rf
	wsRouter := RouterGroupApp.Ws
	roleRouter := RouterGroupApp.Role

	MainGroup := r.Group("/v1/api")
	{
		MainGroup.GET("/check-status", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		}) // track monitor
	}
	{
		userRouter.InitUserRouter(MainGroup)
		authRouter.InitAuthRouter(MainGroup)
		rfRouter.InitRfRouter(MainGroup)
		wsRouter.InitWebSocketRouter(MainGroup)
		roleRouter.Role.InitRoleRouter(MainGroup)
		RouterGroupApp.Channel.InitChannelRouter(MainGroup)
		RouterGroupApp.Connection.InitConnectionRouter(MainGroup)
		RouterGroupApp.Group.InitGroupRouter(MainGroup)
		RouterGroupApp.Message.InitMessageRouter(MainGroup)
		RouterGroupApp.Upload.InitUploadRouter(MainGroup)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
