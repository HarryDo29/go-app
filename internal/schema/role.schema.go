package schema

import (
	"time"

	"go-app/global"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CollectionNameRole = "roles"

// DbRole representing role collection in MongoDB
type DbRole struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoleName  string             `bson:"role_name" json:"role_name"`
	RoleNote  string             `bson:"role_note" json:"role_note"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

func init() {
	global.RegisterMongoCollection(CollectionNameRole)
}
