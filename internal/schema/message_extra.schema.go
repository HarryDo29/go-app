package schema

import (
	"go-app/global"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CollectionNameMessageExtras = "message_extras"

// MessageExtras representing collection message_extras
// saves each message states (like read, delivered status) by each user
type MessageExtras struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"uid" json:"uid"` // user id
	ChannelID primitive.ObjectID `bson:"channel_id" json:"channel_id"`
	MsgID     primitive.ObjectID `bson:"msg_id" json:"msg_id"`
	Version   int64              `bson:"version" json:"version"`
	Sync      bool               `bson:"sync" json:"sync"` // sync state
}

func init() {
	global.RegisterMongoCollection(CollectionNameMessageExtras)
}
