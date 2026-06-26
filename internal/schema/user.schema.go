package schema

import (
	"time"

	"go-app/global"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CollectionNameUser = "users"

// DbUser representing user collection in MongoDB
type DbUser struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserName  string             `bson:"user_name" json:"user_name"`
	Password  string             `bson:"password" json:"password"`
	Email     string             `bson:"email" json:"email"`
	AvatarUrl string             `bson:"avatar_url" json:"avatar_url"`
	IsActive  bool               `bson:"is_active" json:"is_active"`
	Role      primitive.ObjectID `bson:"role" json:"role"` // Reference to Role ObjectID list
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

func init() {
	global.RegisterMongoCollection(CollectionNameUser)
}
