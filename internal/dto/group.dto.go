package dto

import (
	"go-app/internal/schema"
)

type CreateGroupDto struct {
	GroupName string `json:"group_name" binding:"required"`
	OwnerId   string `json:"owner_id"  binding:"required"`
	// MemberIds   []string `json:"member_ids" binding:"required"`
	MemberCount int64 `json:"member_count"  binding:"required"`
}

type UpdateGroupDto struct {
	GroupName   string             `json:"group_name"`
	MemberCount int64              `json:"member_count"`
	Status      schema.GroupStatus `json:"status"`
}

type GroupResponseDto struct {
	GroupId     string                      `json:"group_id"`
	GroupName   string                      `json:"group_name"`
	OwnerId     string                      `json:"owner_id"`
	MemberCount int64                       `json:"member_count"`
	Status      schema.GroupStatus          `json:"status"`
	Members     *[]ChannelMemberResponseDto `json:"members"`
}
