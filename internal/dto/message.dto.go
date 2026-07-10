package dto

// ─── Message ──────────────────────────────────────────────────────────────────

type CreateMessageDto struct {
	ChannelId      string `json:"channel_id" binding:"required"`
	Content        string `json:"content" binding:"required"`
	MsgType        string `json:"msg_type" binding:"required"`
	RepliedToMsgId string `json:"replied_to_msg_id,omitempty"`
}

type UpdateMessageDto struct {
	Content string `json:"content"`
	Status  string `json:"status"`
}

type MessageResponseDto struct {
	MsgId          string `json:"msg_id"`
	ChannelId      string `json:"channel_id"`
	FromId         string `json:"from_id"`
	Content        string `json:"content"`
	MsgType        string `json:"msg_type"`
	MsgSeq         int64  `json:"msg_seq"`
	Status         string `json:"status"`
	IsDelete       bool   `json:"is_delete"`
	RepliedToMsgId string `json:"replied_to_msg_id,omitempty"`
	CreatedAt      string `json:"created_at"`
}
