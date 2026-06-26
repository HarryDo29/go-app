package schema

import (
	"time"

	"go-app/global"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CollectionNameGroup = "groups"

type GroupStatus string

const (
	GroupStatusActive  GroupStatus = "ACTIVE"
	GroupStatusDeleted GroupStatus = "DELETED"
)

// Group representing collection group
type DbGroup struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	OwnerID     primitive.ObjectID `bson:"owner_id" json:"owner_id"`
	MemberCount int64              `bson:"member_count" json:"member_count"`
	Status      GroupStatus        `bson:"status" json:"status"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt   *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

func init() {
	global.RegisterMongoCollection(CollectionNameGroup)
}
