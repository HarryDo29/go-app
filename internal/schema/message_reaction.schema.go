package schema

import (
	"time"

	"go-app/global"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CollectionNameMessReaction = "message_reactions"

// MessReaction representing collection mess_reaction
type MessageReaction struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MsgID          primitive.ObjectID `bson:"msg_id" json:"msg_id"`
	TypeOfReaction string             `bson:"type_of_reaction" json:"type_of_reaction"`
	CreatedBy      primitive.ObjectID `bson:"created_by" json:"created_by"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
}

func init() {
	global.RegisterMongoCollection(CollectionNameMessReaction)
}
