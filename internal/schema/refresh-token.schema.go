package schema

import (
	"time"

	"go-app/global"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CollectionNameRefreshToken = "refresh_tokens"

// DbRefreshToken representing refresh token collection in MongoDB
type DbRefreshToken struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"` // Reference to User ObjectID
	Token     string             `bson:"token" json:"token"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

func init() {
	global.RegisterMongoCollection(CollectionNameRefreshToken)
}
