package mapper

import (
	dto "go-app/internal/dto"
	"go-app/internal/schema"
)

// ToGroupResponseDto maps DbGroup database schema model to GroupResponseDto
func ToGroupResponseDto(group *schema.DbGroup, members *[]dto.ChannelMemberResponseDto) *dto.GroupResponseDto {
	if group == nil {
		return nil
	}
	return &dto.GroupResponseDto{
		GroupId:     group.ID.Hex(),
		GroupName:   group.Name,
		OwnerId:     group.OwnerID.Hex(),
		MemberCount: group.MemberCount,
		Status:      group.Status,
		Members:     members,
	}
}
