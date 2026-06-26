package message

import (
	dto "go-app/internal/dto"
	"go-app/internal/websocket"
	"go-app/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MessageController struct {
	messageService IMessageService
	hub            *websocket.Hub
}

func NewMessageController(
	messageService IMessageService,
	hub *websocket.Hub,
) *MessageController {
	return &MessageController{
		messageService: messageService,
		hub:            hub,
	}
}

// POST /api/messages
func (mc *MessageController) CreateMessage(c *gin.Context) {
	var msgDto dto.CreateMessageDto
	if err := c.ShouldBindJSON(&msgDto); err != nil {
		response.ErrorResponse(c, response.ErrCodeCreateFailed)
		return
	}

	userId := c.GetString("user-id")
	if userId == "" {
		response.ErrorResponse(c, response.ErrCodeAuthFailed)
		return
	}

	msg, err := mc.messageService.CreateMessage(userId, msgDto)
	if err != nil {
		response.ErrorResponse(c, response.ErrCodeCreateFailed)
		return
	}

	// Bắn realtime qua WebSocket
	// mc.hub.Broadcast(websocket.MessagePayload{
	// 	ChannelId: msg.ChannelID.Hex(),
	// 	SenderId:  msg.FromID.Hex(),
	// 	Content:   msg.Content,
	// 	Type:      string(msg.MsgType),
	// })

	response.SuccessResponse(c, response.ErrCodeSuccess, msg)
}

// PUT /api/messages/:id
func (mc *MessageController) UpdateMessage(c *gin.Context) {
	msgId := c.Param("message-id")
	if msgId == "" {
		response.ErrorResponse(c, response.ErrCodeBodyInvalid)
		return
	}
	userId := c.GetString("user-id")
	if userId == "" {
		response.ErrorResponse(c, response.ErrCodeAuthFailed)
		return
	}

	var req dto.UpdateMessageDto
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, response.ErrCodeBodyInvalid)
		return
	}

	msg, err := mc.messageService.UpdateMessage(msgId, userId, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Có thể thêm event Broadcast update_message nếu cần

	c.JSON(http.StatusOK, msg)
}

// DELETE /api/messages/:id/recall
func (mc *MessageController) RecallMessage(c *gin.Context) {
	msgId := c.Param("message-id")
	userId := c.GetString("user-id")

	ok := mc.messageService.RecallMessage(msgId, userId)
	if !ok {
		response.ErrorResponse(c, response.ErrCodeDeleteFailed)
		return
	}

	// Có thể thêm event Broadcast recall_message nếu cần
	response.SuccessResponse(c, response.ErrCodeSuccess, gin.H{"delete": true})
}

// POST /api/messages/:id/hide
func (mc *MessageController) HideMessageForMe(c *gin.Context) {
	msgId := c.Param("message-id")
	userId := c.GetString("user-id")

	// Yêu cầu client truyền channel_id qua query string
	channelId := c.Query("channel_id")
	if channelId == "" {
		response.ErrorResponse(c, response.ErrCodeParamInvalid)
		return
	}

	err := mc.messageService.HideMessageForMe(msgId, userId, channelId)
	if err != nil {
		response.ErrorResponse(c, response.ErrCodeCreateFailed)
		return
	}

	response.SuccessResponse(c, response.ErrCodeSuccess, gin.H{"message": "success"})
}

// DELETE /api/channels/:channelId/history
func (mc *MessageController) DeleteChatHistory(c *gin.Context) {
	channelId := c.Param("channel-id")
	userId := c.GetString("user-id")

	// Lấy seq query nếu có truyền, mặc định là 0 (xoá đến tin nhắn hiện tại)
	upToSeqStr := c.DefaultQuery("up_to_seq", "0")
	upToSeq, _ := strconv.ParseInt(upToSeqStr, 10, 64)

	err := mc.messageService.DeleteChatHistory(userId, channelId, upToSeq)
	if err != nil {
		response.ErrorResponse(c, response.ErrCodeCreateFailed)
		return
	}

	response.SuccessResponse(c, response.ErrCodeSuccess, gin.H{"message": "success"})
}

func (mc *MessageController) GetMessageByID(c *gin.Context) {
	msgId := c.Param("message-id")

	msg, err := mc.messageService.GetMessageByID(msgId)
	if err != nil {
		response.ErrorResponse(c, response.ErrCodeGetFailed)
		return
	}

	if msg == nil {
		response.ErrorResponse(c, response.ErrCodeNotFound)
		return
	}

	response.SuccessResponse(c, response.ErrCodeSuccess, msg)
}

// GET /api/channels/:channelId/messages
func (mc *MessageController) GetMessagesByChannel(c *gin.Context) {
	channelId := c.Param("channel-id")
	userId := c.GetString("user-id")

	limitStr := c.DefaultQuery("limit", "20")
	beforeSeqStr := c.DefaultQuery("before_seq", "0")

	limit, _ := strconv.ParseInt(limitStr, 10, 64)
	beforeSeq, _ := strconv.ParseInt(beforeSeqStr, 10, 64)

	msgs, err := mc.messageService.GetMessagesByChannel(channelId, userId, limit, beforeSeq)
	if err != nil {
		response.ErrorResponse(c, response.ErrCodeNotFound)
		return
	}

	if msgs == nil {
		msgs = &[]dto.MessageResponseDto{} // Trả về mảng rỗng thay vì null
	}

	response.SuccessResponse(c, response.ErrCodeSuccess, msgs)
}
