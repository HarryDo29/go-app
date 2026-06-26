package dto

// ─── Message Extra (ẩn tin nhắn phía cá nhân) ─────────────────────────────────

type CreateMessageExtraDto struct {
	UserId string `json:"user_id" binding:"required"`
	ChannelId string `json:"channel_id" binding:"required"`
	MsgId string `json:"msg_id" binding:"required"`
	Version   int64  `json:"version"`
}

type UpdateMessageExtraDto struct {
	Version int64 `json:"version"`
	Sync    bool  `json:"sync"`
}

type MessageExtraResponseDto struct {
	ExtraId string `json:"extra_id"`
	UserId string `json:"user_id"`
	ChannelId string `json:"channel_id"`
	MsgId string `json:"msg_id"`
	Version   int64  `json:"version"`
	Sync      bool   `json:"sync"`
}
