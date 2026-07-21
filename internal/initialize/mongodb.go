package initialize

import (
	"context"
	"go-app/global"
	_ "go-app/internal/schema"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func InitMongoDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	m := global.Config.Mongo

	//
	clientOptions := options.Client().ApplyURI(m.URI)

	// kết nối với mongodb
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		global.Logger.Error("Init MongoDB connection failed: ", zap.Error(err))
		panic(err)
	}

	// kiểm tra kết nối với mongodb
	if err := client.Ping(ctx, nil); err != nil {
		global.Logger.Error("Ping MongoDB failed: ", zap.Error(err))
		panic(err)
	}
	// chọn database
	db := client.Database(m.Database)

	// Gán vào biến global để sử dụng toàn hệ thống
	global.Mgo = &global.MongoDB{
		Client:   client,
		Database: db,
	}

	// Tự động kiểm tra và tạo mới các collection nếu chưa tồn tại
	// collections, err := db.ListCollectionNames(ctx, bson.D{})
	// if err == nil {
	// 	existing := make(map[string]bool)
	// 	for _, name := range collections {
	// 		existing[name] = true
	// 	}

	// 	for _, name := range global.MongoCollectionsToCreate {
	// 		if !existing[name] {
	// 			// Tạo mới collection một cách tường minh
	// 			err = db.CreateCollection(ctx, name)
	// 			if err != nil {
	// 				global.Logger.Error("Create MongoDB collection failed: ", zap.String("collection", name), zap.Error(err))
	// 			}
	// 		}
	// 	}
	// } else {
	// 	global.Logger.Error("List MongoDB collection names failed: ", zap.Error(err))
	// }

	global.Logger.Info("Init MongoDB connection success")
}
