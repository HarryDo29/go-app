//go:build wireinject

package wire

import (
	"go-app/internal/channel"
	channelRepo "go-app/internal/channel/repo"
	connection "go-app/internal/connection"
	"go-app/internal/group"
	messageRepo "go-app/internal/message/repo"
	"go-app/internal/role"
	"go-app/internal/user"

	"github.com/google/wire"
)

func InitGroupRouterHandler() (*group.GroupController, error) {
	wire.Build(
		role.NewRoleRepo,
		user.NewUserRepo,
		user.NewUserService,
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
		group.NewGroupService,
		group.NewGroupController,
	)
	return new(group.GroupController), nil
}
