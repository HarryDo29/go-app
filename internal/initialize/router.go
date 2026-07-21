package initialize

import (
	routers "go-app/internal/router"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	return routers.NewRouter()
}
