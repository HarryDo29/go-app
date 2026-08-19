package websocket

import "encoding/json"

const (
	// --- Channel ---
	EventNewChannel         = "NEW_CHANNEL"
	EventUpdatedChannel     = "UPDATED_CHANNEL"
	EventDeletedChannel     = "DELETED_CHANNEL"
	EventRemovedFromChannel = "REMOVED_FROM_CHANNEL"
	// --- Message ---
	EventNewMessage     = "NEW_MESSAGE"
	EventUpdatedMessage = "UPDATED_MESSAGE"
	EventRecallMessage  = "RECALLED_MESSAGE"
	// --- Connection(Friend) ---
	EventNewConnection = "NEW_CONNECTION"
	// --- UserTyping---
	EventUserTyping     = "USER_TYPING"
	EventUserStopTyping = "USER_STOP_TYPING"
)

type ClientMessagePayload struct {
	Event     string `json:"event"` // "UNREGISTER"...
	ChannelId string `json:"channelId"`
	UserId    string `json:"userId"`
}

// WsResponse là struct chung để bắn dữ liệu realtime từ Server --> Client
// FE nhận event và tự phân loại dựa vào field Event
type WsResponse struct {
	Event   string      `json:"event"`   // "NEW_MESSAGE", "NEW_NOTIFICATION", "GROUP_CREATED"...
	Payload interface{} `json:"payload"` // nội dung: bất kỳ struct nào
}

// WsRequest là struct để nhận dữ liệu realtime từ Client --> Server
// FE gửi event và payload đến Server thông qua struct này
type WsRequest struct {
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload"` // Keeps payload as raw JSON bytes
}

type TypingPayload struct {
	ChannelId string `json:"channelId"`
	UserId    string `json:"userId"`
}
