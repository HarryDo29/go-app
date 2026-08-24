package websocket

import (
	"context"
	"fmt"
	"go-app/global"
	"log"
	"sync"
	"time"
)

var ctx = context.Background()
var timeDuration = 60 * time.Second

type typingState struct {
	cancel chan struct{} // Channel to signal "cancel this goroutine"
}

type GetChannelMembersFunc func(channelId string) []string

type Hub struct {
	mu sync.RWMutex // lock protect clients
	// map[UserId] map[ConnectionId] *Client
	// support 1 user can use multi-devices in a time
	clients     map[string]map[string]*Client
	typingUsers map[string]map[string]*typingState

	register   chan *Client        // khi user online (kết nối) --> đưa channel register để xử lý
	unregister chan *Client        // khi user ngắt kết nối --> đưa vào channel unregister để xử lý
	typing     chan *TypingPayload // khi user typing --> đưa vào channel typing để xử lý
	stopTyping chan *TypingPayload // khi user stop typing --> đưa vào channel stopTyping để xử lý
	ping       chan *Client        // when you ping --> update last seen in redis

	GetChannelMembersFunc GetChannelMembersFunc
}

func NewHub() *Hub {
	return &Hub{
		clients:     make(map[string]map[string]*Client),
		typingUsers: make(map[string]map[string]*typingState),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		typing:      make(chan *TypingPayload),
		stopTyping:  make(chan *TypingPayload),
		ping:        make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case payload := <-h.typing:
			h.handleUserTyping(payload.ChannelId, payload.UserId)

		case payload := <-h.stopTyping:
			h.handleUserStopTyping(payload.ChannelId, payload.UserId)

		case payload := <-h.ping:
			h.handlePing(payload.UserId)
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) HandleTyping(payload *TypingPayload) {
	h.typing <- payload
}

func (h *Hub) Ping(client *Client) {
	h.ping <- client
}

//-----------------------------------------------------------------

func getUserPresenceKey(userId string) string {
	return fmt.Sprintf("presence:user:%s", userId)
}

func (h *Hub) registerClient(client *Client) {
	if client == nil || client.UserId == "" {
		return
	}

	if h.clients[client.UserId] == nil {
		h.clients[client.UserId] = make(map[string]*Client)
	}

	h.clients[client.UserId][client.ConnectionId] = client
	global.Rdb.HSet(ctx, getUserPresenceKey(client.UserId), "1", timeDuration)
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

	global.Rdb.HDel(ctx, getUserPresenceKey(client.UserId), "1")
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

// BroadcastToUserIds notifies all active member users in a channel
func (h *Hub) BroadcastToUserIds(userIds []string, event string, payload interface{}) {
	for _, userId := range userIds {
		if !h.IsOnline(userId) {
			continue
		}

		h.SendToUser(userId, WsResponse{
			Event:   event,
			Payload: payload,
		})
	}
}

func (h *Hub) handleUserTyping(channelId string, userId string) {
	if state, ok := h.typingUsers[channelId][userId]; ok {
		close(state.cancel)
	}

	cancelChannel := make(chan struct{})
	h.typingUsers[channelId][userId] = &typingState{
		cancel: cancelChannel,
	}

	// 2. (Optional) Broadcast "USER_TYPING" event to other channel members
	// h.BroadcastToUserIds(memberUserIds, "USER_TYPING", TypingPayload{ChannelId: channelId, UserId: userId})
	if h.GetChannelMembersFunc != nil {
		memberIds := h.GetChannelMembersFunc(channelId)
		for _, memberId := range memberIds {
			if memberId == userId {
				continue // Skip sender
			}
			h.SendToUser(memberId, WsResponse{
				Event:   EventUserTyping,
				Payload: TypingPayload{ChannelId: channelId, UserId: userId},
			})
		}
	}

	go func() {
		select {
		case <-time.After(3 * time.Second):
			h.stopTyping <- &TypingPayload{
				ChannelId: channelId,
				UserId:    userId,
			}
		case <-cancelChannel:
			return
		}
	}()
}

func (h *Hub) handleUserStopTyping(channelId string, userId string) {
	// 1. Remove user from typing state map
	if users, ok := h.typingUsers[channelId]; ok {
		delete(users, userId)
		// Clean up empty channel map
		if len(users) == 0 {
			delete(h.typingUsers, channelId)
		}
	}

	// 2. (Optional) Broadcast "USER_STOP_TYPING" event to other channel members
	// h.BroadcastToUserIds(memberUserIds, "USER_STOP_TYPING", TypingPayload{ChannelId: channelId, UserId: userId})
	if h.GetChannelMembersFunc != nil {
		memberIds := h.GetChannelMembersFunc(channelId)
		for _, memberId := range memberIds {
			if memberId == userId {
				continue // Skip sender
			}
			h.SendToUser(memberId, WsResponse{
				Event:   EventUserStopTyping,
				Payload: TypingPayload{ChannelId: channelId, UserId: userId},
			})
		}
	}
}

func (h *Hub) handlePing(userId string) {
	key := getUserPresenceKey(userId)
	// check key exist in redis ?
	if value, err := global.Rdb.ExpireXX(ctx, key, timeDuration).Result(); err == nil && value {
		return
	}

	//	1. Đặt key là userId, value là false (online)
	if err := global.Rdb.HSet(ctx, key, "1", timeDuration).Err(); err != nil {
		log.Println("error set presence user:", err)
	}

	//	2. Bắn event PONG về cho client gửi ping
	h.SendToUser(userId, WsResponse{
		Event: "PONG",
		Payload: map[string]string{
			"msg": "pong",
		},
	})
}

// ---------------------- Admin func ------------------------------------------
// DisconnectUser forces disconnect for all connections of a specific user (e.g. Banned / Logout)
func (h *Hub) DisconnectUser(userId string, reason string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	connections, ok := h.clients[userId]
	if !ok || len(connections) == 0 {
		return false
	}

	res := WsResponse{
		Event: EventForceDisconnect,
		Payload: map[string]string{
			"reason": reason,
		},
	}

	for _, client := range connections {
		// Send force disconnect notice to client before closing
		select {
		case client.Send <- res:
		default:
		}
		close(client.Send)
		client.Conn.Close()
	}

	delete(h.clients, userId)
	return true
}

// BroadcastSystemAnnouncement sends a real-time message to ALL active connected users
func (h *Hub) BroadcastSystemAnnouncement(title string, message string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	res := WsResponse{
		Event: EventSystemAnnouncement,
		Payload: map[string]string{
			"title":   title,
			"message": message,
		},
	}

	for _, connections := range h.clients {
		for _, client := range connections {
			select {
			case client.Send <- res:
				count++
			default:
			}
		}
	}
	return count
}

// SendAdminAlert sends a targeted warning or notification popup to a single user
func (h *Hub) SendAdminAlert(userId string, title string, message string) bool {
	return h.SendToUser(userId, WsResponse{
		Event: EventAdminAlert,
		Payload: map[string]string{
			"title":   title,
			"message": message,
		},
	})
}

// GetOnlineStats retrieves active user counts and connection metrics for Admin Dashboard
func (h *Hub) GetOnlineStats(includeUserIDs bool) AdminOnlineStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	totalUsers := len(h.clients)
	totalConns := 0
	userIds := make([]string, 0, totalUsers)

	for uID, conns := range h.clients {
		totalConns += len(conns)
		if includeUserIDs {
			userIds = append(userIds, uID)
		}
	}

	return AdminOnlineStats{
		TotalOnlineUsers: totalUsers,
		TotalConnections: totalConns,
		OnlineUserIds:    userIds,
	}
}
