package utils

import (
	"encoding/base64"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EncodeCursor mã hóa ID sang Base64
func EncodeCursor(id primitive.ObjectID) string {
	return base64.StdEncoding.EncodeToString([]byte(id.Hex()))
}

// DecodeCursor giải mã token từ client
func DecodeCursor(encoded string) (primitive.ObjectID, error) {
	bytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return primitive.ObjectIDFromHex(string(bytes))
}
