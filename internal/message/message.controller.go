package message

import (
	dto "go-app/internal/dto"
	"go-app/internal/websocket"
	"go-app/pkg/response"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IWsHub interface {
	Notify(userId string, event string, payload interface{}) bool
}

type MessageController struct {
	messageService IMessageService
	hub            IWsHub
}

func NewMessageController(
	messageService IMessageService,
	hub IWsHub,
) *MessageController {
	return &MessageController{
		messageService: messageService,
		hub:            hub,
	}
}

// POST /api/messages
// CreateMessage godoc
// @Summary      Create message
// @Description  Create a new message in a channel
// @Tags         message
// @Accept       json
// @Produce      json
// @Param        req body dto.CreateMessageDto true "Message Info"
// @Success      200 {object} map[string]interface{}
// @Router       /message [post]
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

	messageResponse := dto.MessageResponseDto{
		MsgId:     msg.ID.Hex(),
		ChannelId: msg.ChannelID.Hex(),
		FromId:    msg.FromID.Hex(),
		Content:   msg.Content,
		MsgType:   string(msg.MsgType),
		MsgSeq:    msg.MsgSeq,
		Status:    string(msg.Status),
		IsDelete:  msg.IsDelete,
		CreatedAt: msg.CreatedAt.Format(time.RFC3339),
	}
	if msg.RepliedToMsgID != primitive.NilObjectID {
		messageResponse.RepliedToMsgId = msg.RepliedToMsgID.Hex()
	}

	// Bắn realtime qua WebSocket đến tất cả member trong channel
	memberIds := mc.messageService.GetMemberIds(msgDto.ChannelId)
	for _, memberId := range memberIds {
		mc.hub.Notify(memberId, websocket.EventNewMessage, messageResponse)
	}

	response.SuccessResponse(c, response.ErrCodeSuccess, msg)
}

// PUT /api/messages/:id
// UpdateMessage godoc
// @Summary      Update message
// @Description  Update message content
// @Tags         message
// @Accept       json
// @Produce      json
// @Param        message-id path string true "Message ID"
// @Param        req body dto.UpdateMessageDto true "Update Message Info"
// @Success      200 {object} map[string]interface{}
// @Router       /message/{message-id} [put]
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

	var updateMsgDto dto.UpdateMessageDto
	if err := c.ShouldBindJSON(&updateMsgDto); err != nil {
		response.ErrorResponse(c, response.ErrCodeBodyInvalid)
		return
	}

	msg, err := mc.messageService.UpdateMessage(msgId, userId, updateMsgDto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Map to MessageResponseDto
	messageResponse := &dto.MessageResponseDto{
		MsgId:     msg.ID.Hex(),
		ChannelId: msg.ChannelID.Hex(),
		FromId:    msg.FromID.Hex(),
		Content:   msg.Content,
		MsgType:   string(msg.MsgType),
		MsgSeq:    msg.MsgSeq,
		Status:    string(msg.Status),
		IsDelete:  msg.IsDelete,
		CreatedAt: msg.CreatedAt.Format(time.RFC3339),
	}
	if msg.RepliedToMsgID != primitive.NilObjectID {
		messageResponse.RepliedToMsgId = msg.RepliedToMsgID.Hex()
	}

	// Broadcast update_message
	memberIds := mc.messageService.GetMemberIds(messageResponse.ChannelId)
	for _, memberId := range memberIds {
		mc.hub.Notify(memberId, websocket.EventUpdatedMessage, messageResponse)
	}

	c.JSON(http.StatusOK, messageResponse)
}

// DELETE /api/messages/:id/recall
// RecallMessage godoc
// @Summary      Recall message
// @Description  Recall a message
// @Tags         message
// @Produce      json
// @Param        message-id path string true "Message ID"
// @Success      200 {object} map[string]interface{}
// @Router       /message/{message-id}/recall [delete]
func (mc *MessageController) RecallMessage(c *gin.Context) {
	msgId := c.Param("message-id")
	userId := c.GetString("user-id")

	msg := mc.messageService.RecallMessage(msgId, userId)
	if msg == nil {
		response.ErrorResponse(c, response.ErrCodeDeleteFailed)
		return
	}

	// Broadcast recall_message
	memberIds := mc.messageService.GetMemberIds(msg.ChannelId)
	for _, memberId := range memberIds {
		mc.hub.Notify(memberId, websocket.EventRecallMessage, msg)
	}

	response.SuccessResponse(c, response.ErrCodeSuccess, gin.H{"msg": "Recalled msg successfully"})
}

// POST /api/messages/:id/hide
// HideMessageForMe godoc
// @Summary      Hide message
// @Description  Hide a message for current user
// @Tags         message
// @Produce      json
// @Param        message-id path string true "Message ID"
// @Param        channel_id query string true "Channel ID"
// @Success      200 {object} map[string]interface{}
// @Router       /message/{message-id}/hide [post]
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
// DeleteChatHistory godoc
// @Summary      Delete chat history
// @Description  Delete chat history in a channel up to a sequence
// @Tags         message
// @Produce      json
// @Param        channel-id path string true "Channel ID"
// @Param        up_to_seq query string false "Up To Sequence"
// @Success      200 {object} map[string]interface{}
// @Router       /channel/{channel-id}/history [delete]
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

// GetMessageByID godoc
// @Summary      Get message by ID
// @Description  Get a single message by ID
// @Tags         message
// @Produce      json
// @Param        message-id path string true "Message ID"
// @Success      200 {object} map[string]interface{}
// @Router       /message/{message-id} [get]
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
// GetMessagesByChannel godoc
// @Summary      Get messages by channel
// @Description  Get messages in a channel with pagination
// @Tags         message
// @Produce      json
// @Param        channel-id path string true "Channel ID"
// @Param        limit query string false "Limit"
// @Param        before_seq query string false "Before Sequence"
// @Success      200 {object} map[string]interface{}
// @Router       /channel/{channel-id}/messages [get]
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
