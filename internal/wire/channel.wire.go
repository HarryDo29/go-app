//go:build wireinject

package wire

import (
	"go-app/internal/channel"
	channelRepo "go-app/internal/channel/repo"
	connection "go-app/internal/connection"
	"go-app/internal/group"
	messageRepo "go-app/internal/message/repo"
	"go-app/internal/user"

	"github.com/google/wire"
)

func InitChannelRouterHandler() (*channel.ChannelController, error) {
	wire.Build(
		channelRepo.NewChannelRepo,
		channelRepo.NewChannelMemberRepo,
		channelRepo.NewChannelUnreadRepo,
		user.NewUserRepo,
		connection.NewConnectionRepo,
		group.NewGroupRepo,
		messageRepo.NewMessageRepo,
		// adapters: bridge local interfaces của channel package
		provideChannelUserRepo,
		provideChannelConnectionRepo,
		provideChannelGroupRepo,
		provideChannelMessageRepo,
		channel.NewChannelService,
		channel.NewChannelController,
	)
	return new(channel.ChannelController), nil
}
