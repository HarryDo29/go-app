//go:build wireinject

package wire

import (
	"go-app/internal/role"
	connection "go-app/internal/connection"
	user "go-app/internal/user"

	"github.com/google/wire"
)

func InitUserRouterHandler() (*user.UserController, error) {
	wire.Build(
		role.NewRoleRepo,
		connection.NewConnectionRepo,
		user.NewUserRepo,
		provideUserRoleRepo,
		provideUserConnectionService,
		user.NewUserService,
		user.NewUserController,
	)
	return new(user.UserController), nil
}
