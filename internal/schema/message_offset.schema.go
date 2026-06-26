package schema

import (
	"go-app/global"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CollectionNameMessageOffsets = "message_offsets"

// MessageOffsets representing collection message_offsets
// saves each message conversation deleted offset by each user
type MessageOffsets struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"uid" json:"uid"` // user id
	ChannelID primitive.ObjectID `bson:"channel_id" json:"channel_id"`
	Offset    int64              `bson:"offset" json:"offset"` // amount of messages deleted
	Version   int64              `bson:"version" json:"version"`
	Sync      bool               `bson:"sync" json:"sync"` // sync state
}

func init() {
	global.RegisterMongoCollection(CollectionNameMessageOffsets)
}
