//go:build wireinject

package wire

import (
	"go-app/internal/channel"
	channelRepo "go-app/internal/channel/repo"
	connection "go-app/internal/connection"
	"go-app/internal/group"
	messageRepo "go-app/internal/message/repo"
	"go-app/internal/user"
	"go-app/internal/websocket"

	"github.com/google/wire"
)

func InitWebSocketHandler() (*websocket.Handler, error) {
	wire.Build(
		user.NewUserRepo,
		channelRepo.NewChannelRepo,
		channelRepo.NewChannelMemberRepo,
		channelRepo.NewChannelUnreadRepo,
		connection.NewConnectionRepo,
		group.NewGroupRepo,
		messageRepo.NewMessageRepo,
		// adapters: bridge local interfaces của channel package
		provideChannelUserRepo,
		provideChannelConnectionRepo,
		provideChannelGroupRepo,
		provideChannelMessageRepo,
		channel.NewChannelService,
		websocket.NewHub,
		websocket.NewHandler,
	)
	return new(websocket.Handler), nil
}
