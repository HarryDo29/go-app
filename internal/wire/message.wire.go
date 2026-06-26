//go:build wireinject

package wire

import (
	channelRepo "go-app/internal/channel/repo"
	"go-app/internal/message"
	messageRepo "go-app/internal/message/repo"
	"go-app/internal/websocket"

	"github.com/google/wire"
)

func InitMessageRouterHandler() (*message.MessageController, error) {
	wire.Build(
		channelRepo.NewChannelRepo,
		channelRepo.NewChannelMemberRepo,
		messageRepo.NewMessageRepo,
		messageRepo.NewMessageOffsetRepo,
		messageRepo.NewMessageExtraRepo,
		// adapter:
		provideMessageChannelRepo,
		provideMessageChannelMemberRepo,
		message.NewMessageService,
		websocket.NewHub,
		message.NewMessageController,
	)
	return new(message.MessageController), nil
}
