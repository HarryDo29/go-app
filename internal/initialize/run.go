package initialize

import (
	"fmt"
	"go-app/global"
	"strconv"
	"time"

	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

func Run() {
	// load config
	LoadConfig()
	fmt.Println("Loading port: ", global.Config.Server.Port)
	// init log
	InitLogger()
	global.Logger.Info("Logger initialized ...", zap.String("ok", "log"))
	// mongo
	InitMongoDB()
	// redis
	InitRedis()
	// init cache
	global.Cache = cache.New(5*time.Minute, 10*time.Minute)
	// miniIO
	InitMinio()
	// websocket
	InitWebSocket()
	// router
	r := InitRouter()
	r.Run(":" + strconv.Itoa(global.Config.Server.Port))
}
