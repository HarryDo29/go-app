package dto

// ─── Message Offset (xóa lịch sử chat) ───────────────────────────────────────

type CreateMessageOffsetDto struct {
	UserId string `json:"user_id" binding:"required"`
	ChannelId string `json:"channel_id" binding:"required"`
	Offset    int64  `json:"offset"`
	Version   int64  `json:"version"`
}

type UpdateMessageOffsetDto struct {
	Offset  int64 `json:"offset"`
	Version int64 `json:"version"`
	Sync    bool  `json:"sync"`
}

type MessageOffsetResponseDto struct {
	OffsetId string `json:"offset_id"`
	UserId string `json:"user_id"`
	ChannelId string `json:"channel_id"`
	Offset    int64  `json:"offset"`
	Version   int64  `json:"version"`
	Sync      bool   `json:"sync"`
}
