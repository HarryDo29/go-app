package user

import (
	"fmt"
	dto "go-app/internal/dto"
	"go-app/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService IUserService
}

func NewUserController(userService IUserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

func (uc *UserController) GetMe(c *gin.Context) {
	userId := c.GetString("user-id")
	fmt.Println("userId:", userId)
	user := uc.userService.GetUserById(userId)
	if user == nil {
		response.ErrorResponse(c, response.ErrCodeUserNotFound)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, user)
}

func (uc *UserController) GetUserById(c *gin.Context) {
	id := c.Param("user-id")
	user := uc.userService.GetUserById(id)
	if user == nil {
		response.ErrorResponse(c, response.ErrCodeUserNotFound)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, user)
}

func (uc *UserController) SearchUsers(c *gin.Context) {
	userId := c.GetString("user-id")
	keyword := c.Query("name")

	if keyword == "" {
		response.SuccessResponse(c, response.ErrCodeSuccess, []interface{}{})
		return
	}

	users := uc.userService.SearchUsers(keyword, userId)
	if users == nil {
		response.ErrorResponse(c, response.ErrCodeNotFound)
		return
	}

	response.SuccessResponse(c, response.ErrCodeSuccess, users)
}

func (uc *UserController) UpdateUser(c *gin.Context) {
	id := c.GetString("user-id")
	var dto dto.UpdateUserDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		response.ErrorResponse(c, response.ErrCodeBodyInvalid)
		return
	}
	user := uc.userService.UpdateUser(id, dto)
	if user == nil {
		response.ErrorResponse(c, response.ErrCodeUpdateFailed)
		return
	}
	response.SuccessResponse(c, response.ErrCodeUserExist, user)
}

func (uc *UserController) DeleteUser(c *gin.Context) {
	id := c.Param("user-id")
	result := uc.userService.DeleteUser(id)
	if !result {
		response.ErrorResponse(c, response.ErrCodeDeleteFailed)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, nil)
}
