package websocket

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
)

type ClientMessagePayload struct {
	Event string `json:"event"` // "UNREGISTER"...
}

// WsResponse là struct chung để bắn dữ liệu realtime từ Server --> Client
// FE nhận event và tự phân loại dựa vào field Event
type WsResponse struct {
	Event   string      `json:"event"`   // "NEW_MESSAGE", "NEW_NOTIFICATION", "GROUP_CREATED"...
	Payload interface{} `json:"payload"` // nội dung: bất kỳ struct nào
}

// Struct riêng cho Thông báo
type NotificationData struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Link  string `json:"link"`
}
