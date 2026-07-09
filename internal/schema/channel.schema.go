package schema

import (
	"time"

	"go-app/global"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CollectionNameChannel = "channels"

type ChannelType string

const (
	ChannelTypeGroup  ChannelType = "group"
	ChannelTypeDirect ChannelType = "direct"
)

// Channel representing collection channel
type DbChannel struct {
	ID             primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	ChannelType    ChannelType          `bson:"channel_type" json:"channel_type"` // group or direct
	ChannelKey     primitive.ObjectID   `bson:"channel_key" json:"channel_key"`   // group_id if group or connection_id if direct
	LastMsgID      primitive.ObjectID   `bson:"last_msg_id" json:"last_msg_id"`
	LastMsgSeq     int64                `bson:"last_msg_seq" json:"last_msg_seq"`
	LastMsgTime    time.Time            `bson:"last_msg_time" json:"last_msg_time"`
	IsActive       bool                 `bson:"is_active" json:"is_active"`
	ParticipantIds []primitive.ObjectID `bson:"participant_ids" json:"participant_ids"`
	CreatedAt      time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time            `bson:"updated_at" json:"updated_at"`
}

func init() {
	global.RegisterMongoCollection(CollectionNameChannel)
}
