package initianlize

import (
	"fmt"
	"go-app/global"
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
	// mysql
	// InitMySql()
	// mongo
	InitMongoDB()
	// redis
	InitRedis()
	// init cache
	global.Cache = cache.New(5*time.Minute, 10*time.Minute)
	// miniIO
	InitMinio()
	// router
	r := InitRouter()
	r.Run(":8081")
}
