package schema

import (
	"time"

	"go-app/global"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CollectionNameConnection = "connections"

type ConnectionStatus string

const (
	ConnectionStatusPending  ConnectionStatus = "PENDING"
	ConnectionStatusAccepted ConnectionStatus = "ACCEPTED"
	ConnectionStatusRejected ConnectionStatus = "REJECTED"
	ConnectionStatusRemoved  ConnectionStatus = "REMOVED"
)

// Connection representing collection connection
type DbConnection struct {
	ID             primitive.ObjectID    `bson:"_id,omitempty" json:"id"`
	RequesterID    primitive.ObjectID    `bson:"requester_id" json:"requester_id"`
	ReceiverID     primitive.ObjectID    `bson:"receiver_id" json:"receiver_id"`
	ParticipantIDs [2]primitive.ObjectID `bson:"participant_ids" json:"participant_ids"` // tránh tạo trùng (có A -> B ròi thì ko cho B -> A)
	Status         ConnectionStatus      `bson:"status" json:"status"`
	CreatedAt      time.Time             `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time             `bson:"updated_at" json:"updated_at"`
	AcceptedAt     *time.Time            `bson:"accepted_at,omitempty" json:"accepted_at,omitempty"`
}

func init() {
	global.RegisterMongoCollection(CollectionNameConnection)
}
