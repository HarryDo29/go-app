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
	GetRefreshTokens(userId string) []schema.DbRefreshToken
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
func (rfp *refreshTokenRepo) GetRefreshTokens(userId string) []schema.DbRefreshToken {
	userID := utils.ObjectIDFromHex(userId)
	if userID == primitive.NilObjectID {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var tokens []schema.DbRefreshToken
	collection := global.Mgo.Database.Collection(schema.CollectionNameRefreshToken)
	cursor, err := collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &tokens); err != nil {
		return nil
	}
	return tokens
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
