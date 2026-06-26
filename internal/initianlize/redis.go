package initianlize

import (
	"context"
	"fmt"
	"go-app/global"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var ctx = context.Background()

func InitRedis() {
	r := global.Config.Redis

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", r.Host, r.Port),
		Password: r.Password, // mặc định ko có password
		DB:       r.DB,
		PoolSize: r.Pool, // Số lượng kết nối tối đa trong pool
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		global.Logger.Error("Init Redis connection failed: ", zap.Error(err))
		panic(err)
	}

	global.Rdb = rdb
	global.Logger.Info("Init Redis connection success")
	redisExample()
}

func redisExample() {
	global.Rdb.Set(ctx, "name", "HarryDo29", 0).Err()
	val, err := global.Rdb.Get(ctx, "name").Result()
	if err != nil {
		global.Logger.Error("Redis set failed: ", zap.Error(err))
	}
	global.Logger.Info("Redis set success: ", zap.String("name", val))

	val, err = global.Rdb.Get(ctx, "name").Result()
	if err != nil {
		global.Logger.Error("Redis get failed: ", zap.Error(err))
	}
	global.Logger.Info("Redis get success: ", zap.String("name", val))
}
