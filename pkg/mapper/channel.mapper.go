package mapper

import (
	dto "go-app/internal/dto"
	"go-app/internal/schema"
)

// ToChannelResponseDto maps DbChannel database schema model to ChannelResponseDto
func ToChannelResponseDto(
	channel *schema.DbChannel,
	subject *dto.UserResponseDto,
	group *dto.GroupResponseDto,
	lastMsg *dto.MessageResponseDto,
) *dto.ChannelResponseDto {
	if channel == nil {
		return nil
	}
	return &dto.ChannelResponseDto{
		ChannelId:   channel.ID.Hex(),
		ChannelType: string(channel.ChannelType),
		ChannelKey:  channel.ChannelKey.Hex(),
		Subject:     subject,
		Group:       group,
		LastMsg:     lastMsg,
		UpdatedAt:   channel.UpdatedAt,
	}
}
