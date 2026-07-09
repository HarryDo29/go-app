//go:build wireinject

package wire

import (
	channelRepo "go-app/internal/channel/repo"
	"go-app/internal/message"
	messageRepo "go-app/internal/message/repo"

	"github.com/google/wire"
)

// InitMessageService khởi tạo MessageService với tất cả dependencies (không bao gồm Hub)
func InitMessageService() (message.IMessageService, error) {
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
	)
	return nil, nil
}
