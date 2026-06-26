package refreshtoken

import (
	"context"
	"go-app/global"
	dto "go-app/internal/dto"
	"go-app/internal/schema"
	"go-app/pkg/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IRefreshTokenRepo interface {
	CreateRefreshToken(rf dto.CreateFreshTokenDto) bool
	GetRefreshToken(userId string) schema.DbRefreshToken
	RemoveRefreshToken(userId string, rfToken string) bool
}

type refreshTokenRepo struct{}

// CreateRefreshToken implements [IRefreshTokenRepo].
func (rfp *refreshTokenRepo) CreateRefreshToken(rf dto.CreateFreshTokenDto) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameRefreshToken)
	token := schema.DbRefreshToken{
		UserID:    rf.UserId,
		Token:     rf.Token,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := collection.InsertOne(ctx, &token)
	return err == nil
}

// GetRefreshToken implements [IRefreshTokenRepo].
func (rfp *refreshTokenRepo) GetRefreshToken(userId string) schema.DbRefreshToken {
	objUserID := utils.ObjectIDFromHex(userId)
	if objUserID == primitive.NilObjectID {
		return schema.DbRefreshToken{}
	}

	var token schema.DbRefreshToken
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameRefreshToken)
	err := collection.FindOne(ctx, bson.M{"user_id": objUserID}).Decode(&token)
	if err == nil {
		return token
	}
	return schema.DbRefreshToken{} // thay cho nil
}

// RemoveRefreshToken implements [IRefreshTokenRepo].
func (rfp *refreshTokenRepo) RemoveRefreshToken(userId string, rfToken string) bool {
	objUserID := utils.ObjectIDFromHex(userId)
	if objUserID == primitive.NilObjectID {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameRefreshToken)
	_, err := collection.DeleteOne(ctx, bson.M{"user_id": objUserID, "token": rfToken})
	return err == nil
}

func NewRefreshTokenRepo() IRefreshTokenRepo {
	return &refreshTokenRepo{}
}
