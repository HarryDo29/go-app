package websocket

type Hub struct {
	// clients lưu trữ: map[UserId] map[ConnectionId] *Client
	// Hỗ trợ 1 user có thể kết nối trên nhiều thiết bị (nhiều tab/app) cùng lúc
	clients map[string]map[string]*Client

	register   chan *Client // khi user online (kết nối) --> đưa channel register để xử lý
	unregister chan *Client // khi user ngắt kết nối --> đưa vào channel unregister để xử lý
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) registerClient(client *Client) {
	if client == nil || client.UserId == "" {
		return
	}

	if h.clients[client.UserId] == nil {
		h.clients[client.UserId] = make(map[string]*Client)
	}

	h.clients[client.UserId][client.ConnectionId] = client
} // done

func (h *Hub) unregisterClient(client *Client) {
	if client == nil {
		return
	}

	connections, ok := h.clients[client.UserId] // lấy danh sách connections của userId
	if ok {
		delete(connections, client.ConnectionId) // xoá connection disconnect

		if len(connections) == 0 {
			delete(h.clients, client.UserId) // nếu không còn kết nối thì off user
		}
	}
	close(client.Send) // đóng channel Send
} // done

func (h *Hub) IsOnline(userId string) bool {
	connections, ok := h.clients[userId]
	return ok && len(connections) > 0
} // done

// SendToUser gửi event WsResponse đến tất cả thiết bị của 1 user
// FE sẽ nhận event và tự phân loại dựa vào field Event
func (h *Hub) SendToUser(userId string, res WsResponse) bool {
	connections, ok := h.clients[userId]
	if !ok || len(connections) == 0 {
		return false
	}

	for _, client := range connections {
		client.Send <- res
	}

	return true
}

// Notify bắn realtime đến 1 user, nhận event và payload riêng lẻ
// Dùng để implement IWsHub interface ở các package khác mà không cần import WsResponse
func (h *Hub) Notify(userId string, event string, payload interface{}) bool {
	return h.SendToUser(userId, WsResponse{
		Event:   event,
		Payload: payload,
	})
}
