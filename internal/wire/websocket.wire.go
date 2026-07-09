//go:build wireinject

package wire

import (
	"go-app/internal/websocket"

	"github.com/google/wire"
)

func InitWebSocketHandler() (*websocket.Handler, error) {
	wire.Build(
		websocket.NewHub,
		websocket.NewHandler,
	)
	return new(websocket.Handler), nil
}
