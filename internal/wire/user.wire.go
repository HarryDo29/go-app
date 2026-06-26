//go:build wireinject

package wire

import (
	"go-app/internal/role"
	user "go-app/internal/user"

	"github.com/google/wire"
)

func InitUserRouterHandler() (*user.UserController, error) {
	wire.Build(
		role.NewRoleRepo,
		user.NewUserRepo,
		user.NewUserService,
		user.NewUserController,
	)
	return new(user.UserController), nil
}
