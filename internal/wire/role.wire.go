//go:build wireinject

package wire

import (
	"go-app/internal/role"

	"github.com/google/wire"
)

func InitRoleRouterHandler() (*role.RoleController, error) {
	wire.Build(
		role.NewRoleRepo,
		role.NewRoleService,
		role.NewRoleController,
	)
	return new(role.RoleController), nil
}
