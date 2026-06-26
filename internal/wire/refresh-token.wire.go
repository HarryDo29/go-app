//go:build wireinject

package wire

import (
	rf "go-app/internal/refresh-token"

	"github.com/google/wire"
)

func InitRefreshTokenRouterHandler() (*rf.RefreshTokenController, error) {
	wire.Build(
		rf.NewRefreshTokenRepo,
		rf.NewRefreshTokenService,
		rf.NewRefreshTokenController,
	)
	return new(rf.RefreshTokenController), nil
}
