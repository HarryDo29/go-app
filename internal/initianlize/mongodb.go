package initianlize

import (
	"context"
	"fmt"
	"go-app/global"
	_ "go-app/internal/schema"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitMongoDB() (*global.MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	m := global.Config.Mongo

	//
	clientOptions := options.Client().ApplyURI(m.URI)

	// kết nối với mongodb
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("connect mongodb failed: %w", err)
	}

	// kiểm tra kết nối với mongodb
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("ping mongodb failed: %w", err)
	}
	// chọn database
	db := client.Database(m.Database)

	// Gán vào biến global để sử dụng toàn hệ thống
	global.Mgo = &global.MongoDB{
		Client:   client,
		Database: db,
	}

	// Tự động kiểm tra và tạo mới các collection nếu chưa tồn tại
	collections, err := db.ListCollectionNames(ctx, bson.D{})
	if err == nil {
		existing := make(map[string]bool)
		for _, name := range collections {
			existing[name] = true
		}

		for _, name := range global.MongoCollectionsToCreate {
			if !existing[name] {
				// Tạo mới collection một cách tường minh
				_ = db.CreateCollection(ctx, name)
			}
		}
	}

	return global.Mgo, nil
}
