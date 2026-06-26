package schema

import (
	"time"

	"go-app/global"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CollectionNameMessage = "messages"

type MessageStatus string
type MessageType string

const (
	MessageStatusSent      MessageStatus = "SENT"
	MessageStatusDelivered MessageStatus = "DELIVERED"
	MessageStatusRead      MessageStatus = "READ"
)
const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
	MessageTypeFile  MessageType = "file"
)

// Message representing collection message
type Message struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ChannelID      primitive.ObjectID `bson:"channel_id" json:"channel_id"`
	FromID         primitive.ObjectID `bson:"from_id" json:"from_id"`
	Content        string             `bson:"content" json:"content"`
	MsgType        MessageType        `bson:"msg_type" json:"msg_type"` // text, image, file, sticker, system, etc.
	MsgSeq         int64              `bson:"msg_seq" json:"msg_seq"`
	Status         MessageStatus      `bson:"status" json:"status"` // sent, delivered, read
	IsDelete       bool               `bson:"is_delete" json:"is_delete"`
	RepliedToMsgID primitive.ObjectID `bson:"replied_to_msg_id,omitempty" json:"replied_to_msg_id,omitempty"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

func init() {
	global.RegisterMongoCollection(CollectionNameMessage)
}
