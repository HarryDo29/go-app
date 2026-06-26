package schema

import (
	"time"

	"go-app/global"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CollectionNameChannelUnread = "channel_unread"

// ChannelUnread representing collection channel_unread
// saves unread channel state by each user
type DbChannelUnread struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"uid"` // user id
	ChannelID   primitive.ObjectID `bson:"channel_id" json:"channel_id"`
	LastMsgID   primitive.ObjectID `bson:"last_msg_id" json:"last_msg_id"`
	LastMsgTime time.Time          `bson:"last_msg_time" json:"last_msg_time"`
	IsActive    bool               `b.son:"is_active" json:"is_active"`
	Unread      int64              `bson:"unread" json:"unread"` // count of unread messages, 0 -> read
	Version     int64              `bson:"version" json:"version"`
}

func init() {
	global.RegisterMongoCollection(CollectionNameChannelUnread)
}
