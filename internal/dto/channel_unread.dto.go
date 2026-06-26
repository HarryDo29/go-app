package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateChannelUnreadDto struct {
	UserIds   []string `json:"user_ids"`
	ChannelId string   `json:"channel_id"`
	Unread    int64    `json:"unread"`
	Version   int64    `json:"version"`
}

type UpdateChannelUnreadDto struct {
	LastMsgID   primitive.ObjectID `json:"last_msg_id"`
	LastMsgTime time.Time          `json:"last_msg_time"`
	Unread      int64              `json:"unread"`
	Version     int64              `json:"version"`
}

type ChannelUnreadResponseDto struct {
	UnreadId string    `json:"unread_id"`
	UserId string    `json:"user_id"`
	ChannelId string    `json:"channel_id"`
	LastMsgId string    `json:"last_msg_id"`
	LastMsgTime time.Time `json:"last_msg_time"`
	Unread      int64     `json:"unread"`
	Version     int64     `json:"version"`
}
