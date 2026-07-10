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

// GetMe godoc
// @Summary      Get current user
// @Description  Get current authenticated user info
// @Tags         user
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Router       /user/me [get]
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

// GetUserById godoc
// @Summary      Get user by ID
// @Description  Get user information by their ID
// @Tags         user
// @Produce      json
// @Param        user-id path string true "User ID"
// @Success      200 {object} map[string]interface{}
// @Router       /user/{user-id} [get]
func (uc *UserController) GetUserById(c *gin.Context) {
	id := c.Param("user-id")
	user := uc.userService.GetUserById(id)
	if user == nil {
		response.ErrorResponse(c, response.ErrCodeUserNotFound)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, user)
}

// SearchUsers godoc
// @Summary      Search users
// @Description  Search users by name
// @Tags         user
// @Produce      json
// @Param        name query string false "Search Keyword"
// @Success      200 {object} map[string]interface{}
// @Router       /user/search [get]
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

// UpdateUser godoc
// @Summary      Update user
// @Description  Update current user information
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        req body dto.UpdateUserDto true "Update User Info"
// @Success      200 {object} map[string]interface{}
// @Router       /user [put]
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

// DeleteUser godoc
// @Summary      Delete user
// @Description  Delete user by ID
// @Tags         user
// @Produce      json
// @Param        user-id path string true "User ID"
// @Success      200 {object} map[string]interface{}
// @Router       /user/{user-id} [delete]
func (uc *UserController) DeleteUser(c *gin.Context) {
	id := c.Param("user-id")
	result := uc.userService.DeleteUser(id)
	if !result {
		response.ErrorResponse(c, response.ErrCodeDeleteFailed)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, nil)
}
