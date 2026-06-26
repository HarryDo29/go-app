package websocket

type Hub struct {
	clients      map[string]map[string]*Client  // map client id với danh sách client
	channels     map[string]map[string]struct{} // map channel với danh sách user trong channel: channel_id: []user_id
	register     chan *Client                   // khi user online (kết nối) --> đưa channel register để xử lý
	unregister   chan *Client                   // khi user ngắt kết nối --> đưa vào channel unregister để xử lý
	joinChannel  chan ChannelPayload
	leaveChannel chan ChannelPayload
	broadcast    chan MessagePayload // khi có tin nhắn --> đưa vào channel broadcast
}

type ChannelPayload struct {
	ChannelId string
	UserId string
}

type ClientMessagePayload struct {
	Event     string `json:"event"`
	ChannelId string `json:"channel_id"`
	Content   string `json:"content"`
	Type      string `json:"type"`
}

type MessagePayload struct {
	ChannelId string `json:"channel_id"`
	SenderId  string `json:"sender_id"`
	Content   string `json:"content"`
	Type      string `json:"type"` // text/image/file
}

func NewHub() *Hub {
	return &Hub{
		clients:      make(map[string]map[string]*Client),
		channels:     make(map[string]map[string]struct{}),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		joinChannel:  make(chan ChannelPayload),
		leaveChannel: make(chan ChannelPayload),
		broadcast:    make(chan MessagePayload),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			// register
			h.registerClient(client)

		case client := <-h.unregister:
			// unregister
			h.unregisterClient(client)

		case payload := <-h.joinChannel:
			// join a channel
			h.addClientToChannel(payload.ChannelId, payload.UserId)

		case payload := <-h.leaveChannel:
			// leave a channel
			h.removeClientFromChannel(payload.ChannelId, payload.UserId)

		case message := <-h.broadcast:
			// broadcast (message)
			h.broadcastToChannel(message)
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) JoinChannel(channelId string, userId string) {
	h.joinChannel <- ChannelPayload{
		ChannelId: channelId,
		UserId:    userId,
	}
}

func (h *Hub) LeaveChannel(channelId string, userId string) {
	h.leaveChannel <- ChannelPayload{
		ChannelId: channelId,
		UserId:    userId,
	}
}

func (h *Hub) Broadcast(message MessagePayload) {
	h.broadcast <- message
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
			delete(h.clients, client.UserId)             // nếu không còn kết nối thì off user
			h.removeClientFromAllChannels(client.UserId) // xoá connection ra khỏi channel
		}
	}
	close(client.Send) // đóng
} // done

func (h *Hub) addClientToChannel(channelId string, userId string) {
	if channelId == "" || userId == "" {
		return
	}

	if h.channels[channelId] == nil {
		h.channels[channelId] = make(map[string]struct{})
	}

	h.channels[channelId][userId] = struct{}{} // đánh dấu user trong channel đang onl
} // done

func (h *Hub) removeClientFromChannel(channelId string, userId string) {
	users, ok := h.channels[channelId]
	if !ok {
		return
	}

	delete(users, userId) // xoá đánh dấu user online trong channel

	if len(users) == 0 {
		delete(h.channels, channelId)
	}
}

func (h *Hub) removeClientFromAllChannels(userId string) {
	for channelId := range h.channels {
		h.removeClientFromChannel(channelId, userId)
	}
} // done

func (h *Hub) broadcastToChannel(message MessagePayload) {
	if message.ChannelId == "" {
		return
	}

	users, ok := h.channels[message.ChannelId]
	if !ok {
		return
	}

	if _, ok := users[message.SenderId]; !ok {
		return
	}

	for userId := range users {
		for _, client := range h.clients[userId] {
			client.Send <- message
		}
	}
} // done

func (h *Hub) IsOnline(userId string) bool {
	connections, ok := h.clients[userId]
	return ok && len(connections) > 0
} // done

func (h *Hub) SendToUser(userId string, message MessagePayload) bool {
	connections, ok := h.clients[userId] // lấy danh sách các connection của 1 user
	if !ok || len(connections) == 0 {
		return false
	}

	for _, client := range connections {
		client.Send <- message
	}

	return true
}
