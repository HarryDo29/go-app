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

func InitGroupService() (group.IGroupService, error) {
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
		provideUserRoleRepo,
		provideUserConnectionService,
		channel.NewChannelService,
		group.NewGroupService,
	)
	return nil, nil
}
