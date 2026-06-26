package conection

import (
	dto "go-app/internal/dto"
	"go-app/pkg/response"

	"github.com/gin-gonic/gin"
)

type ConnectionController struct {
	connectionService IConnectionService
}

func NewConnectionController(connectionService IConnectionService) *ConnectionController {
	return &ConnectionController{
		connectionService: connectionService,
	}
}

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
	response.SuccessResponse(c, response.ErrCodeSuccess, result)
}

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

func (cc *ConnectionController) GetConnectionByUserId(c *gin.Context) {
	userId := c.GetString("user-id")
	if userId == "" {
		response.ErrorResponse(c, response.ErrCodeParamInvalid)
		return
	}
	connection := cc.connectionService.GetConnectionByUserId(userId)
	if connection == nil {
		response.ErrorResponse(c, response.ErrCodeNotFound)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, connection)
}

func (cc *ConnectionController) AcceptedConnection(c *gin.Context) {
	id := c.Param("connection-id")
	if id == "" {
		response.ErrorResponse(c, response.ErrCodeParamInvalid)
		return
	}
	connection := cc.connectionService.AcceptedConnection(id)
	if connection == nil {
		response.ErrorResponse(c, response.ErrCodeUpdateFailed)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, connection)
}

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
