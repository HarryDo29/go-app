//go:build wireinject

package wire

import (
	"go-app/internal/auth"
	rf "go-app/internal/refresh-token"
	"go-app/internal/role"
	user "go-app/internal/user"

	"github.com/google/wire"
)

func InitAuthRouterHandler() (*auth.AuthController, error) {
	wire.Build(
		user.NewUserRepo,
		rf.NewRefreshTokenRepo,
		rf.NewRefreshTokenService,
		role.NewRoleRepo,
		auth.NewAuthService,
		auth.NewAuthController,
	)
	return new(auth.AuthController), nil
}
