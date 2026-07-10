package conection

import (
	dto "go-app/internal/dto"
	"go-app/internal/schema"
	"go-app/internal/websocket"
	"go-app/pkg/response"

	"github.com/gin-gonic/gin"
)

type IWsHub interface {
	Notify(userId string, event string, payload interface{}) bool
}

type ConnectionController struct {
	connectionService IConnectionService
	hub               IWsHub
}

// CreateConnection godoc
// @Summary      Create connection
// @Description  Create a new connection/friend request
// @Tags         connection
// @Accept       json
// @Produce      json
// @Param        req body dto.ConnectionDto true "Connection Info"
// @Success      200 {object} map[string]interface{}
// @Router       /connection [post]
func (cc *ConnectionController) CreateConnection(c *gin.Context) {
	var conDto dto.ConnectionDto
	if err := c.ShouldBindJSON(&conDto); err != nil {
		return
	}
	// tạo connection cùng với channel, members, unreads
	result := cc.connectionService.CreateConnection(conDto)
	if result == nil {
		response.ErrorResponse(c, response.ErrCodeCreateFailed)
		return
	}

	// Gửi realtime báo có Friend Request mới
	cc.hub.Notify(conDto.ReceiverId, websocket.EventNewConnection, result.Connection)
	// response
	response.SuccessResponse(c, response.ErrCodeSuccess, result.Connection)
}

// GetConnection godoc
// @Summary      Get connection details
// @Description  Get connection details by participants
// @Tags         connection
// @Accept       json
// @Produce      json
// @Param        req body dto.Participants true "Participants Info"
// @Success      200 {object} map[string]interface{}
// @Router       /connection/detail [post]
func (cc *ConnectionController) GetConnection(c *gin.Context) {
	var participants dto.Participants
	if err := c.ShouldBindJSON(&participants); err != nil {
		response.ErrorResponse(c, response.ErrCodeBodyInvalid)
		return
	}
	connection := cc.connectionService.GetConnection(participants.Participants)
	if connection == nil {
		response.ErrorResponse(c, response.ErrCodeNotFound)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, connection)
}

// GetConnectionByUserId godoc
// @Summary      Get user connections
// @Description  Get connections for the current user, optionally filtered by status
// @Tags         connection
// @Produce      json
// @Param        status query string false "Connection Status"
// @Success      200 {object} map[string]interface{}
// @Router       /connection/user [get]
func (cc *ConnectionController) GetConnectionByUserId(c *gin.Context) {
	userId := c.GetString("user-id")
	status := c.Query("status")
	if userId == "" {
		response.ErrorResponse(c, response.ErrCodeParamInvalid)
		return
	}
	connection := cc.connectionService.GetConnectionByUserId(userId, status)
	if connection == nil {
		response.ErrorResponse(c, response.ErrCodeNotFound)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, connection)
}

// RespondConnection godoc
// @Summary      Respond to connection request
// @Description  Accept or reject a connection request
// @Tags         connection
// @Produce      json
// @Param        connection-id path string true "Connection ID"
// @Param        status query string true "Status (accepted, rejected)"
// @Success      200 {object} map[string]interface{}
// @Router       /connection/{connection-id}/respond [put]
func (cc *ConnectionController) RespondConnection(c *gin.Context) {
	connectionId := c.Param("connection-id")
	if connectionId == "" {
		response.ErrorResponse(c, response.ErrCodeParamInvalid)
		return
	}

	status := schema.ConnectionStatus(c.Query("status"))
	if status == "" {
		response.ErrorResponse(c, response.ErrCodeParamInvalid)
		return
	}

	switch status {
	case schema.ConnectionStatusRejected:
		success := cc.connectionService.RejectedConnection(connectionId)
		if !success {
			response.ErrorResponse(c, response.ErrCodeUpdateFailed)
			return
		}
		response.SuccessResponse(c, response.ErrCodeSuccess, "Connection rejected")
		return

	case schema.ConnectionStatusAccepted:
		connection := cc.connectionService.AcceptedConnection(connectionId)
		if connection == nil {
			response.ErrorResponse(c, response.ErrCodeUpdateFailed)
			return
		}
		// Gửi realtime báo có Chat Channel mới cho cả 2 người
		reqChannel := *connection.Channel
		reqChannel.Subject = connection.Receiver

		recChannel := *connection.Channel
		recChannel.Subject = connection.Requester

		cc.hub.Notify(connection.Connection.RequesterId, websocket.EventNewChannel, reqChannel)
		cc.hub.Notify(connection.Connection.ReceiverId, websocket.EventNewChannel, recChannel)

		response.SuccessResponse(c, response.ErrCodeSuccess, recChannel)
		return
	}

	response.ErrorResponse(c, response.ErrCodeParamInvalid)
}

// DeleteConnection godoc
// @Summary      Delete connection
// @Description  Delete a connection by ID
// @Tags         connection
// @Produce      json
// @Param        connection-id path string true "Connection ID"
// @Success      200 {object} map[string]interface{}
// @Router       /connection/{connection-id} [delete]
func (cc *ConnectionController) DeleteConnection(c *gin.Context) {
	id := c.Param("connection-id")
	if id == "" {
		response.ErrorResponse(c, response.ErrCodeParamInvalid)
		return
	}
	connection := cc.connectionService.DeleteConnection(id)
	if connection == false {
		response.ErrorResponse(c, response.ErrCodeUpdateFailed)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, connection)
}

func NewConnectionController(connectionService IConnectionService, hub IWsHub) *ConnectionController {
	return &ConnectionController{
		connectionService: connectionService,
		hub:               hub,
	}
}
