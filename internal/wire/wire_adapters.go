// wire_adapters.go
// File này chứa các hàm adapter để giải quyết vấn đề không khớp kiểu (type mismatch)
// giữa các interface cục bộ (local interfaces) và các interface từ các package khác khi dùng Google Wire.
// Vì Wire so khớp dependency dựa trên tên kiểu dữ liệu cụ thể (Name-based Matching) chứ không dựa trên
// cấu trúc phương thức (Structural/Duck Typing), chúng ta cần các hàm adapter này làm cầu nối trung gian.
package wire

import (
	"go-app/internal/channel"
	channelRepo "go-app/internal/channel/repo"
	connection "go-app/internal/connection"
	"go-app/internal/group"
	"go-app/internal/message"
	messageRepo "go-app/internal/message/repo"
	"go-app/internal/user"
)

func provideChannelUserRepo(r user.IUserRepo) channel.IUserRepo { return r }

func provideChannelConnectionRepo(r connection.IConnectionRepo) channel.IConnectionRepo { return r }

func provideChannelGroupRepo(r group.IGroupRepo) channel.IGroupRepo { return r }

func provideChannelMessageRepo(r messageRepo.IMessageRepo) channel.IMessageRepo { return r }

func provideMessageChannelRepo(r channelRepo.IChannelRepo) message.IChannelRepo { return r }

func provideMessageChannelMemberRepo(r channelRepo.IChannelMemberRepo) message.IChannelMemberRepo {
	return r
}
