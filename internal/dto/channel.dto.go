package dto

import "time"

type CreateChannelDto struct {
	ChannelType string `json:"channel_type" binding:"required"`
	ChannelKey  string `json:"channel_key" binding:"required"`
}

type UpdateChannelDto struct {
	LastMsgId   string    `json:"last_msg_id"`
	LastMsgSeq  int64     `json:"last_msg_seq"`
	LastMsgTime time.Time `json:"last_msg_time"`
}

type ChannelResponseDto struct {
	ChannelId string              `json:"channel_id"`
	ChannelType string              `json:"channel_type"`
	ChannelKey  string              `json:"channel_key"`
	Subject     *UserResponseDto    `json:"subject"`
	Group       *GroupResponseDto   `json:"group"`
	LastMsg     *MessageResponseDto `json:"last_msg"`
}
