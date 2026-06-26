package utils

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ObjectIDFromHex converts a hex string to primitive.ObjectID.
// If the hex string is invalid, it returns primitive.NilObjectID.
func ObjectIDFromHex(hex string) primitive.ObjectID {
	id, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		return primitive.NilObjectID
	}
	return id
}
