package schema

import (
	"time"

	"go-app/global"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CollectionNameChannelMembers = "channel_members"

type ChannelMemberStatus string

const (
	ChannelMemberStatusActive ChannelMemberStatus = "active"
	ChannelMemberStatusLeft   ChannelMemberStatus = "left"
	ChannelMemberStatusKicked ChannelMemberStatus = "kicked"
)

type ChannelMemberRole string

const (
	ChannelMemberRoleAdmin  ChannelMemberRole = "admin"
	ChannelMemberCoAdmin    ChannelMemberRole = "co-admin"
	ChannelMemberRoleMember ChannelMemberRole = "member"
)

// ChannelMembers representing collection channel_members
type DbChannelMember struct {
	ID        primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	ChannelID primitive.ObjectID  `bson:"channel_id" json:"channel_id"`
	UserID    primitive.ObjectID  `bson:"user_id" json:"user_id"`
	Role      ChannelMemberRole   `bson:"role" json:"role"`     // e.g. admin, member
	Status    ChannelMemberStatus `bson:"status" json:"status"` // active | left | kicked
	JoinedAt  time.Time           `bson:"joined_at" json:"joined_at"`
	LeftAt    *time.Time          `bson:"left_at,omitempty" json:"left_at,omitempty"`
}

func init() {
	global.RegisterMongoCollection(CollectionNameChannelMembers)
}
