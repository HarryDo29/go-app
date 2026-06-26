package dto

import "time"

type CreateChannelMemberDto struct {
	ChannelId string   `json:"channel_id"`
	UserIds []string `json:"user_ids"`
	Role      string   `json:"role"`
	Status    string   `json:"status"`
}

type UpdateChannelMemberDto struct {
	Role   string `json:"role"`
	Status string `json:"status"`
}

type ChannelMemberResponseDto struct {
	MemberId string           `json:"member_id"`
	ChannelId string           `json:"channel_id"`
	User      *UserResponseDto `json:"user"`
	Role      string           `json:"role"`
	Status    string           `json:"status"`
	JoinedAt  time.Time        `json:"joined_at"`
}
