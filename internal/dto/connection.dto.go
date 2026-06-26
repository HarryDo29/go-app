package dto

import "time"

type ConnectionDto struct {
	RequesterId string `json:"request_id"`
	ReceiverId string `json:"receive_id"`
}

type Participants struct {
	Participants [2]string `json:"participants"`
}

type ConnectionResponseDto struct {
	ConnectionId string     `json:"connection_id"`
	RequesterId string     `json:"requester_id"`
	ReceiverId string     `json:"receiver_id"`
	ParticipantIDs [2]string  `json:"participant_ids"`
	Status         string     `json:"status"`
	AcceptedAt     *time.Time `json:"accepted_at,omitempty"`
}

type CreateConnectionResponseDto struct {
	Connection *ConnectionResponseDto      `json:"connection"`
	Channel    *ChannelResponseDto         `json:"channel"`
	Members    *[]ChannelMemberResponseDto `json:"members"`
	Unreads    bool                        `json:"unreads"`
}
