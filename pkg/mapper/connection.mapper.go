package mapper

import (
	dto "go-app/internal/dto"
	"go-app/internal/schema"
)

// ToConnectionResponseDto maps DbConnection database schema model to ConnectionResponseDto
func ToConnectionResponseDto(conn *schema.DbConnection) *dto.ConnectionResponseDto {
	if conn == nil {
		return nil
	}
	participants := [2]string{
		conn.ParticipantIDs[0].Hex(),
		conn.ParticipantIDs[1].Hex(),
	}
	return &dto.ConnectionResponseDto{
		ConnectionId:   conn.ID.Hex(),
		RequesterId:    conn.RequesterID.Hex(),
		ReceiverId:     conn.ReceiverID.Hex(),
		ParticipantIDs: participants,
		Status:         string(conn.Status),
		AcceptedAt:     conn.AcceptedAt,
	}
}

// ToConnectionResponseDtoList maps a slice of DbConnection to a slice of ConnectionResponseDto
func ToConnectionResponseDtoList(connections []schema.DbConnection) []dto.ConnectionResponseDto {
	if connections == nil {
		return []dto.ConnectionResponseDto{}
	}
	result := make([]dto.ConnectionResponseDto, len(connections))
	for i, conn := range connections {
		if res := ToConnectionResponseDto(&conn); res != nil {
			result[i] = *res
		}
	}
	return result
}
