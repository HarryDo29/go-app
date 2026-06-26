package conection

import (
	"context"
	"fmt"
	"go-app/global"
	dto "go-app/internal/dto"
	"go-app/internal/schema"
	"go-app/pkg/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IConnectionRepo interface {
	CreateConnection(conDto dto.ConnectionDto) *schema.DbConnection
	GetConnectionById(id primitive.ObjectID) *schema.DbConnection
	GetConnection(participantIDs [2]primitive.ObjectID) *schema.DbConnection
	GetConnectionByUserId(userId string) *[]schema.DbConnection
	AcceptedConnection(id primitive.ObjectID) *schema.DbConnection
	DeleteConnection(id primitive.ObjectID) bool
}

type ConnectionRepo struct{}

// CreateConnection implements [IConnectionRepo].
func (c *ConnectionRepo) CreateConnection(conDto dto.ConnectionDto) *schema.DbConnection {
	requesterId := utils.ObjectIDFromHex(conDto.RequesterId)
	if requesterId == primitive.NilObjectID {
		return nil
	}
	receiverId := utils.ObjectIDFromHex(conDto.ReceiverId)
	if receiverId == primitive.NilObjectID {
		return nil
	}
	participantIDs := [2]primitive.ObjectID{requesterId, receiverId}
	connection := c.GetConnection(participantIDs)
	if connection != nil {
		return connection
	}

	nConnection := schema.DbConnection{
		ID:             primitive.NewObjectID(),
		Status:         schema.ConnectionStatusPending,
		RequesterID:    requesterId,
		ReceiverID:     receiverId,
		ParticipantIDs: participantIDs,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		AcceptedAt:     nil,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameConnection)
	_, err := collection.InsertOne(ctx, &nConnection)
	if err != nil {
		fmt.Println("Connection create failed: ", err)
		return nil
	}
	return &nConnection
}

// GetConnectionById implements [IConnectionRepo].
func (c *ConnectionRepo) GetConnectionById(id primitive.ObjectID) *schema.DbConnection {
	var connection schema.DbConnection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameConnection)
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&connection)
	if err != nil {
		fmt.Println("Connection get failed: ", err)
		return nil
	}

	return &connection
}

// GetConnection implements [IConnectionRepo].
func (c *ConnectionRepo) GetConnection(participantIDs [2]primitive.ObjectID) *schema.DbConnection {
	var connection schema.DbConnection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameConnection)
	err := collection.FindOne(ctx, bson.M{"participant_ids": participantIDs}).Decode(&connection)
	if err != nil {
		fmt.Println("Connection get failed: ", err)
		return nil
	}

	return &connection
}

// GetConnectionByUserId implements [IConnectionRepo].
func (c *ConnectionRepo) GetConnectionByUserId(userId string) *[]schema.DbConnection {
	var connections []schema.DbConnection

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameConnection)
	userID := utils.ObjectIDFromHex(userId)
	filter := bson.M{
		"participant_ids": userID,
		// "status":          schema.ConnectionStatusAccepted,
	}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		fmt.Println("Connection get failed: ", err)
		return nil
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &connections); err != nil {
		fmt.Errorf("Find connections of user: %s failed", userID)
		return nil
	}

	return &connections
}

// AcceptedConnection implements [IConnectionRepo].
func (c *ConnectionRepo) AcceptedConnection(id primitive.ObjectID) *schema.DbConnection {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameConnection)
	_, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"status":      schema.ConnectionStatusAccepted,
				"accepted_at": time.Now(),
			},
		},
	)
	if err != nil {
		fmt.Println("Connection accepted failed: ", err)
		return nil
	}
	return c.GetConnectionById(id)
}

// DeleteConnection implements [IConnectionRepo].
func (c *ConnectionRepo) DeleteConnection(id primitive.ObjectID) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameConnection)
	_, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"status":     schema.ConnectionStatusRemoved,
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		fmt.Println("Connection delete failed: ", err)
		return false
	}
	return true
}

func NewConnectionRepo() IConnectionRepo {
	return &ConnectionRepo{}
}
