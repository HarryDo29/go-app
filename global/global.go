package global

import (
	"go-app/pkg/logger"
	"go-app/pkg/setting"
	"go-app/internal/websocket"

	"github.com/minio/minio-go/v7"
	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoDB struct {
	Client   *mongo.Client   // dùng cho nghiệp vụ phức tạp
	Database *mongo.Database // thao tác vs database chỉ định (CRUD)
}

var (
	Config setting.Config    // config local
	Logger *logger.LoggerZap // logger
	// Mdb    *gorm.DB          // mysql
	Rdb *redis.Client // redis
	Mgo *MongoDB      // mongodb connection wrapper
	// ModelsToMigrate          []interface{}     // model auto migrate
	MongoCollectionsToCreate []string // mongodb collections to create
	Minio                    *minio.Client
	Cache                    *cache.Cache
	WsHub                    *websocket.Hub // singleton Hub dùng chung toàn app
)

// RegisterModel function registers a model to the list of models to migrate
// func RegisterModel(model interface{}) {
// 	ModelsToMigrate = append(ModelsToMigrate, model)
// }

// RegisterMongoCollection registers a collection name to be created on startup
func RegisterMongoCollection(name string) {
	MongoCollectionsToCreate = append(MongoCollectionsToCreate, name)
}
