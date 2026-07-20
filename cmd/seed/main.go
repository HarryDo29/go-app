package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"time"

	"go-app/global"
	"go-app/internal/initianlize"
	"go-app/internal/schema"
	"go-app/pkg/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const defaultPassword = "Root123@"

type seedConfig struct {
	users                 int
	groups                int
	directChannels        int
	connectionsPerUser    int
	historyDays           int
	minGroupsPerUser      int
	maxGroupsPerUser      int
	minMessagesPerUser    int
	maxMessagesPerUser    int
	appendMessagesPerUser int
	reset                 bool
}

type groupSeed struct {
	group        schema.DbGroup
	channel      schema.DbChannel
	participants []primitive.ObjectID
}

type directSeed struct {
	connection   schema.DbConnection
	channel      schema.DbChannel
	participants []primitive.ObjectID
}

func main() {
	cfg := seedConfig{}
	flag.IntVar(&cfg.users, "users", 500, "number of fake users to create, excluding the root/admin user")
	flag.IntVar(&cfg.groups, "groups", 50, "number of fake groups to create")
	flag.IntVar(&cfg.directChannels, "direct-channels", 0, "number of accepted direct connections/channels to create; 0 derives from connections-per-user")
	flag.IntVar(&cfg.connectionsPerUser, "connections-per-user", 10, "minimum accepted direct connections for each user")
	flag.IntVar(&cfg.historyDays, "history-days", 10, "number of previous days to spread each user's messages across")
	flag.IntVar(&cfg.minGroupsPerUser, "min-groups-per-user", 10, "minimum groups each user joins")
	flag.IntVar(&cfg.maxGroupsPerUser, "max-groups-per-user", 15, "maximum groups each user joins")
	flag.IntVar(&cfg.minMessagesPerUser, "min-messages-per-user", 50, "minimum total messages created by each user")
	flag.IntVar(&cfg.maxMessagesPerUser, "max-messages-per-user", 80, "maximum total messages created by each user")
	flag.IntVar(&cfg.appendMessagesPerUser, "append-messages-per-user", 0, "append this many new messages for each existing user without recreating other data")
	flag.BoolVar(&cfg.reset, "reset", false, "drop seeded collections before inserting")
	flag.Parse()

	if cfg.users < 1 || cfg.groups < 1 || cfg.directChannels < 0 || cfg.connectionsPerUser < 0 || cfg.historyDays < 1 || cfg.minGroupsPerUser < 0 || cfg.maxGroupsPerUser < 0 || cfg.minMessagesPerUser < 0 || cfg.maxMessagesPerUser < 0 || cfg.appendMessagesPerUser < 0 {
		log.Fatal("invalid seed size")
	}
	if cfg.connectionsPerUser >= cfg.users+1 {
		cfg.connectionsPerUser = cfg.users
	}
	if cfg.minGroupsPerUser > cfg.groups {
		cfg.minGroupsPerUser = cfg.groups
	}
	if cfg.maxGroupsPerUser > cfg.groups {
		cfg.maxGroupsPerUser = cfg.groups
	}
	if cfg.maxGroupsPerUser < cfg.minGroupsPerUser {
		cfg.maxGroupsPerUser = cfg.minGroupsPerUser
	}
	if cfg.maxMessagesPerUser < cfg.minMessagesPerUser {
		cfg.maxMessagesPerUser = cfg.minMessagesPerUser
	}

	initianlize.LoadConfig()
	initianlize.InitMongoDB()
	defer func() {
		_ = global.Mgo.Client.Disconnect(context.Background())
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	db := global.Mgo.Database

	if cfg.appendMessagesPerUser > 0 {
		if err := appendMessagesForExistingUsers(ctx, db, rng, cfg.appendMessagesPerUser); err != nil {
			log.Fatalf("append messages: %v", err)
		}
		return
	}

	if cfg.reset {
		if err := resetCollections(ctx, db); err != nil {
			log.Fatalf("reset collections: %v", err)
		}
	}

	userRoleID := mustEnsureRole(ctx, db, "user", "normal")
	adminRoleID := mustEnsureRole(ctx, db, "admin", "highest previllege")
	passwordHash, err := utils.HashPassword(defaultPassword)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	users := buildUsers(cfg.users, userRoleID, adminRoleID, passwordHash)
	if err := upsertUsers(ctx, db, users); err != nil {
		log.Fatalf("upsert users: %v", err)
	}

	userIDs := make([]primitive.ObjectID, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	groups := buildGroups(cfg, rng, userIDs)
	if err := insertGroups(ctx, db, groups); err != nil {
		log.Fatalf("insert groups: %v", err)
	}

	directs := buildDirectConnections(cfg, rng, userIDs)
	if err := insertDirectConnections(ctx, db, directs); err != nil {
		log.Fatalf("insert direct connections: %v", err)
	}

	allChannels := make([]schema.DbChannel, 0, len(groups)+len(directs))
	for _, group := range groups {
		allChannels = append(allChannels, group.channel)
	}
	for _, direct := range directs {
		allChannels = append(allChannels, direct.channel)
	}
	if err := insertChannels(ctx, db, allChannels); err != nil {
		log.Fatalf("insert channels: %v", err)
	}

	channelMembers := buildChannelMembers(groups, directs)
	if err := insertMany(ctx, db.Collection(schema.CollectionNameChannelMembers), docsFromChannelMembers(channelMembers), 1000); err != nil {
		log.Fatalf("insert channel members: %v", err)
	}

	messages, channelLast := buildMessages(cfg, rng, groups, directs)
	if err := insertMany(ctx, db.Collection(schema.CollectionNameMessage), docsFromMessages(messages), 1000); err != nil {
		log.Fatalf("insert messages: %v", err)
	}
	if err := updateChannelLastMessages(ctx, db, channelLast); err != nil {
		log.Fatalf("update channel last messages: %v", err)
	}

	unreads := buildChannelUnreads(rng, groups, directs, channelLast)
	if err := insertMany(ctx, db.Collection(schema.CollectionNameChannelUnread), docsFromUnreads(unreads), 1000); err != nil {
		log.Fatalf("insert channel unread: %v", err)
	}

	offsets := buildMessageOffsets(groups, directs)
	if err := insertMany(ctx, db.Collection(schema.CollectionNameMessageOffsets), docsFromOffsets(offsets), 1000); err != nil {
		log.Fatalf("insert message offsets: %v", err)
	}

	extras := buildMessageExtras(rng, messages, channelParticipants(groups, directs))
	if err := insertMany(ctx, db.Collection(schema.CollectionNameMessageExtras), docsFromExtras(extras), 2000); err != nil {
		log.Fatalf("insert message extras: %v", err)
	}

	reactions := buildMessageReactions(rng, messages, channelParticipants(groups, directs))
	if err := insertMany(ctx, db.Collection(schema.CollectionNameMessReaction), docsFromReactions(reactions), 1000); err != nil {
		log.Fatalf("insert message reactions: %v", err)
	}

	fmt.Printf("Seed completed\n")
	fmt.Printf("- users: %d fake + root@example.com, password: %s\n", cfg.users, defaultPassword)
	fmt.Printf("- groups: %d\n", len(groups))
	fmt.Printf("- direct channels: %d\n", len(directs))
	fmt.Printf("- channels: %d\n", len(allChannels))
	fmt.Printf("- per-user direct connections: at least %d\n", cfg.connectionsPerUser)
	fmt.Printf("- per-user group memberships: %d-%d\n", cfg.minGroupsPerUser, cfg.maxGroupsPerUser)
	fmt.Printf("- message history: %d days, %d-%d total messages/user\n", cfg.historyDays, cfg.minMessagesPerUser, cfg.maxMessagesPerUser)
	fmt.Printf("- channel members: %d\n", len(channelMembers))
	fmt.Printf("- messages: %d\n", len(messages))
	fmt.Printf("- channel unread rows: %d\n", len(unreads))
	fmt.Printf("- message offsets: %d\n", len(offsets))
	fmt.Printf("- message extras: %d\n", len(extras))
	fmt.Printf("- message reactions: %d\n", len(reactions))
}

func resetCollections(ctx context.Context, db *mongo.Database) error {
	collections := []string{
		schema.CollectionNameUser,
		schema.CollectionNameRole,
		schema.CollectionNameGroup,
		schema.CollectionNameConnection,
		schema.CollectionNameChannel,
		schema.CollectionNameChannelMembers,
		schema.CollectionNameChannelUnread,
		schema.CollectionNameMessage,
		schema.CollectionNameMessageExtras,
		schema.CollectionNameMessageOffsets,
		schema.CollectionNameMessReaction,
		schema.CollectionNameRefreshToken,
	}
	for _, name := range collections {
		if err := db.Collection(name).Drop(ctx); err != nil {
			return err
		}
	}
	return nil
}

func mustEnsureRole(ctx context.Context, db *mongo.Database, roleName, note string) primitive.ObjectID {
	now := time.Now()
	filter := bson.M{"role_name": roleName}
	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":        primitive.NewObjectID(),
			"role_name":  roleName,
			"created_at": now,
		},
		"$set": bson.M{
			"role_note":  note,
			"updated_at": now,
		},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var saved schema.DbRole
	if err := db.Collection(schema.CollectionNameRole).FindOneAndUpdate(ctx, filter, update, opts).Decode(&saved); err != nil {
		log.Fatalf("ensure role %s: %v", roleName, err)
	}
	return saved.ID
}

func buildUsers(count int, userRoleID, adminRoleID primitive.ObjectID, passwordHash string) []schema.DbUser {
	now := time.Now()
	users := []schema.DbUser{
		{
			ID:        primitive.NewObjectID(),
			UserName:  "Root Admin",
			Password:  passwordHash,
			Email:     "root@example.com",
			AvatarUrl: avatarURL(0),
			IsActive:  true,
			Role:      adminRoleID,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	firstNames := []string{"An", "Binh", "Chi", "Dung", "Giang", "Ha", "Hieu", "Khanh", "Lan", "Linh", "Long", "Minh", "Nam", "Ngoc", "Phuc", "Quan", "Son", "Thao", "Trang", "Vy"}
	lastNames := []string{"Nguyen", "Tran", "Le", "Pham", "Hoang", "Huynh", "Phan", "Vu", "Vo", "Dang"}
	for i := 1; i <= count; i++ {
		name := fmt.Sprintf("%s %s %03d", firstNames[i%len(firstNames)], lastNames[i%len(lastNames)], i)
		users = append(users, schema.DbUser{
			ID:        primitive.NewObjectID(),
			UserName:  name,
			Password:  passwordHash,
			Email:     fmt.Sprintf("fake.user.%04d@example.com", i),
			AvatarUrl: avatarURL(i),
			IsActive:  true,
			Role:      userRoleID,
			CreatedAt: now.Add(-time.Duration(i%90) * 24 * time.Hour),
			UpdatedAt: now,
		})
	}
	return users
}

func upsertUsers(ctx context.Context, db *mongo.Database, users []schema.DbUser) error {
	models := make([]mongo.WriteModel, 0, len(users))
	for _, user := range users {
		models = append(models, mongo.NewUpdateOneModel().
			SetFilter(bson.M{"email": user.Email}).
			SetUpdate(bson.M{
				"$setOnInsert": bson.M{
					"_id":        user.ID,
					"created_at": user.CreatedAt,
				},
				"$set": bson.M{
					"user_name":  user.UserName,
					"password":   user.Password,
					"email":      user.Email,
					"avatar_url": user.AvatarUrl,
					"is_active":  user.IsActive,
					"role":       user.Role,
					"updated_at": user.UpdatedAt,
				},
			}).
			SetUpsert(true))
	}
	_, err := db.Collection(schema.CollectionNameUser).BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
	return err
}

func buildGroups(cfg seedConfig, rng *rand.Rand, userIDs []primitive.ObjectID) []groupSeed {
	now := time.Now()
	groups := make([]groupSeed, 0, cfg.groups)
	for i := 0; i < cfg.groups; i++ {
		groupID := primitive.NewObjectID()
		channelID := primitive.NewObjectID()
		createdAt := now.Add(-time.Duration(i%30) * 24 * time.Hour)
		groups = append(groups, groupSeed{
			group: schema.DbGroup{
				ID:        groupID,
				Name:      fmt.Sprintf("Seed Group %03d", i+1),
				Status:    schema.GroupStatusActive,
				CreatedAt: createdAt,
				UpdatedAt: now,
			},
			channel: schema.DbChannel{
				ID:          channelID,
				ChannelType: schema.ChannelTypeGroup,
				ChannelKey:  groupID,
				IsActive:    true,
				CreatedAt:   createdAt,
				UpdatedAt:   now,
			},
		})
	}

	participantSets := make([]map[primitive.ObjectID]bool, cfg.groups)
	for i := range participantSets {
		participantSets[i] = make(map[primitive.ObjectID]bool)
	}
	for _, userID := range userIDs {
		groupCount := randomBetween(rng, cfg.minGroupsPerUser, cfg.maxGroupsPerUser)
		for _, groupIndex := range rng.Perm(cfg.groups)[:groupCount] {
			participantSets[groupIndex][userID] = true
		}
	}
	for i := range participantSets {
		if len(participantSets[i]) == 0 {
			participantSets[i][userIDs[rng.Intn(len(userIDs))]] = true
		}

		participants := make([]primitive.ObjectID, 0, len(participantSets[i]))
		for userID := range participantSets[i] {
			participants = append(participants, userID)
		}
		sort.Slice(participants, func(a, b int) bool {
			return participants[a].Hex() < participants[b].Hex()
		})
		ownerID := participants[rng.Intn(len(participants))]
		groups[i].participants = participants
		groups[i].group.OwnerID = ownerID
		groups[i].group.MemberCount = int64(len(participants))
		groups[i].channel.ParticipantIds = participants
	}
	return groups
}

func insertGroups(ctx context.Context, db *mongo.Database, groups []groupSeed) error {
	docs := make([]any, 0, len(groups))
	for _, group := range groups {
		docs = append(docs, group.group)
	}
	return insertMany(ctx, db.Collection(schema.CollectionNameGroup), docs, 1000)
}

func buildDirectConnections(cfg seedConfig, rng *rand.Rand, userIDs []primitive.ObjectID) []directSeed {
	now := time.Now()
	userCount := len(userIDs)
	maxPairs := len(userIDs) * (len(userIDs) - 1) / 2
	target := cfg.directChannels
	minRequired := (userCount*cfg.connectionsPerUser + 1) / 2
	if target < minRequired {
		target = minRequired
	}
	if target > maxPairs {
		target = maxPairs
	}

	seen := make(map[string]bool, target)
	directs := make([]directSeed, 0, target)
	degrees := make(map[primitive.ObjectID]int, len(userIDs))

	appendDirect := func(a, b primitive.ObjectID) bool {
		if a == b || len(directs) >= target {
			return false
		}
		pair := sortedPair(a, b)
		key := pair[0].Hex() + ":" + pair[1].Hex()
		if seen[key] {
			return false
		}
		seen[key] = true

		connectionID := primitive.NewObjectID()
		channelID := primitive.NewObjectID()
		acceptedAt := now.Add(-time.Duration(rng.Intn(30)) * 24 * time.Hour)
		connection := schema.DbConnection{
			ID:             connectionID,
			RequesterID:    a,
			ReceiverID:     b,
			ParticipantIDs: [2]primitive.ObjectID{pair[0], pair[1]},
			Status:         schema.ConnectionStatusAccepted,
			CreatedAt:      acceptedAt.Add(-time.Hour),
			UpdatedAt:      now,
			AcceptedAt:     &acceptedAt,
		}
		channel := schema.DbChannel{
			ID:             channelID,
			ChannelType:    schema.ChannelTypeDirect,
			ChannelKey:     connectionID,
			IsActive:       true,
			ParticipantIds: []primitive.ObjectID{pair[0], pair[1]},
			CreatedAt:      connection.CreatedAt,
			UpdatedAt:      now,
		}
		directs = append(directs, directSeed{connection: connection, channel: channel, participants: channel.ParticipantIds})
		degrees[pair[0]]++
		degrees[pair[1]]++
		return true
	}

	for offset := 1; offset <= cfg.connectionsPerUser/2; offset++ {
		for i, userID := range userIDs {
			appendDirect(userID, userIDs[(i+offset)%userCount])
		}
	}

	for hasUserBelowDegree(userIDs, degrees, cfg.connectionsPerUser) && len(directs) < target {
		for _, userID := range userIDs {
			if degrees[userID] >= cfg.connectionsPerUser {
				continue
			}
			for attempts := 0; attempts < userCount*2 && degrees[userID] < cfg.connectionsPerUser; attempts++ {
				otherID := userIDs[rng.Intn(userCount)]
				appendDirect(userID, otherID)
			}
		}
	}

	for len(directs) < target {
		a := userIDs[rng.Intn(len(userIDs))]
		b := userIDs[rng.Intn(len(userIDs))]
		appendDirect(a, b)
	}
	return directs
}

func insertDirectConnections(ctx context.Context, db *mongo.Database, directs []directSeed) error {
	docs := make([]any, 0, len(directs))
	for _, direct := range directs {
		docs = append(docs, direct.connection)
	}
	return insertMany(ctx, db.Collection(schema.CollectionNameConnection), docs, 1000)
}

func insertChannels(ctx context.Context, db *mongo.Database, channels []schema.DbChannel) error {
	docs := make([]any, 0, len(channels))
	for _, channel := range channels {
		docs = append(docs, channel)
	}
	return insertMany(ctx, db.Collection(schema.CollectionNameChannel), docs, 1000)
}

func buildChannelMembers(groups []groupSeed, directs []directSeed) []schema.DbChannelMember {
	now := time.Now()
	members := make([]schema.DbChannelMember, 0)
	for _, group := range groups {
		for i, userID := range group.participants {
			role := schema.ChannelMemberRoleMember
			if i == 0 {
				role = schema.ChannelMemberRoleAdmin
			} else if i == 1 {
				role = schema.ChannelMemberCoAdmin
			}
			members = append(members, schema.DbChannelMember{
				ID:        primitive.NewObjectID(),
				ChannelID: group.channel.ID,
				UserID:    userID,
				Role:      role,
				Status:    schema.ChannelMemberStatusActive,
				JoinedAt:  group.group.CreatedAt.Add(time.Duration(i) * time.Minute),
			})
		}
	}
	for _, direct := range directs {
		for _, userID := range direct.participants {
			members = append(members, schema.DbChannelMember{
				ID:        primitive.NewObjectID(),
				ChannelID: direct.channel.ID,
				UserID:    userID,
				Role:      schema.ChannelMemberRoleMember,
				Status:    schema.ChannelMemberStatusActive,
				JoinedAt:  now.Add(-24 * time.Hour),
			})
		}
	}
	return members
}

func buildMessages(cfg seedConfig, rng *rand.Rand, groups []groupSeed, directs []directSeed) ([]schema.Message, map[primitive.ObjectID]schema.Message) {
	now := time.Now()
	participantsByChannel := channelParticipants(groups, directs)
	channelsByUser := make(map[primitive.ObjectID][]primitive.ObjectID)
	for channelID, participants := range participantsByChannel {
		for _, userID := range participants {
			channelsByUser[userID] = append(channelsByUser[userID], channelID)
		}
	}

	estimatedMessages := len(channelsByUser) * ((cfg.minMessagesPerUser + cfg.maxMessagesPerUser) / 2)
	messagesByChannel := make(map[primitive.ObjectID][]schema.Message, len(participantsByChannel))
	for userID, channelIDs := range channelsByUser {
		targetMessages := randomBetween(rng, cfg.minMessagesPerUser, cfg.maxMessagesPerUser)
		for i := 0; i < targetMessages; i++ {
			channelID := channelIDs[rng.Intn(len(channelIDs))]
			createdAt := randomMessageTime(rng, now, cfg.historyDays)
			updatedAt := createdAt.Add(time.Duration(rng.Intn(10*60)) * time.Second)
			if updatedAt.After(now) {
				updatedAt = now
			}
			content, msgType := messagePayload(rng, i+1)
			messagesByChannel[channelID] = append(messagesByChannel[channelID], schema.Message{
				ID:        primitive.NewObjectID(),
				ChannelID: channelID,
				FromID:    userID,
				Content:   content,
				MsgType:   msgType,
				Status:    messageStatus(rng),
				IsDelete:  false,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			})
		}
	}

	messages := make([]schema.Message, 0, estimatedMessages)
	lastByChannel := make(map[primitive.ObjectID]schema.Message, len(messagesByChannel))
	for channelID, channelMessages := range messagesByChannel {
		sort.Slice(channelMessages, func(i, j int) bool {
			return channelMessages[i].CreatedAt.Before(channelMessages[j].CreatedAt)
		})
		for i := range channelMessages {
			channelMessages[i].MsgSeq = int64(i + 1)
			messages = append(messages, channelMessages[i])
			lastByChannel[channelID] = channelMessages[i]
		}
	}
	return messages, lastByChannel
}

func updateChannelLastMessages(ctx context.Context, db *mongo.Database, lastByChannel map[primitive.ObjectID]schema.Message) error {
	models := make([]mongo.WriteModel, 0, len(lastByChannel))
	for channelID, msg := range lastByChannel {
		models = append(models, mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": channelID}).
			SetUpdate(bson.M{"$set": bson.M{
				"last_msg_id":   msg.ID,
				"last_msg_seq":  msg.MsgSeq,
				"last_msg_time": msg.CreatedAt,
				"updated_at":    time.Now(),
			}}))
	}
	if len(models) == 0 {
		return nil
	}
	_, err := db.Collection(schema.CollectionNameChannel).BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
	return err
}

func appendMessagesForExistingUsers(ctx context.Context, db *mongo.Database, rng *rand.Rand, messagesPerUser int) error {
	userIDs, err := loadExistingUserIDs(ctx, db)
	if err != nil {
		return err
	}
	channelsByUser, participantsByChannel, err := loadExistingChannelMemberships(ctx, db)
	if err != nil {
		return err
	}
	channelSeqs, err := loadExistingChannelSequences(ctx, db)
	if err != nil {
		return err
	}

	messages, channelLast := buildAppendedMessages(rng, userIDs, channelsByUser, channelSeqs, messagesPerUser)
	if err := insertMany(ctx, db.Collection(schema.CollectionNameMessage), docsFromMessages(messages), 1000); err != nil {
		return err
	}
	if err := updateChannelLastMessages(ctx, db, channelLast); err != nil {
		return err
	}
	if err := updateExistingChannelUnreads(ctx, db, rng, channelLast, participantsByChannel); err != nil {
		return err
	}

	extras := buildMessageExtras(rng, messages, participantsByChannel)
	if err := insertMany(ctx, db.Collection(schema.CollectionNameMessageExtras), docsFromExtras(extras), 2000); err != nil {
		return err
	}

	reactions := buildMessageReactions(rng, messages, participantsByChannel)
	if err := insertMany(ctx, db.Collection(schema.CollectionNameMessReaction), docsFromReactions(reactions), 1000); err != nil {
		return err
	}

	fmt.Printf("Append messages completed\n")
	fmt.Printf("- users: %d\n", len(userIDs))
	fmt.Printf("- appended messages/user: %d\n", messagesPerUser)
	fmt.Printf("- appended messages: %d\n", len(messages))
	fmt.Printf("- touched channels: %d\n", len(channelLast))
	fmt.Printf("- message extras: %d\n", len(extras))
	fmt.Printf("- message reactions: %d\n", len(reactions))
	return nil
}

func loadExistingUserIDs(ctx context.Context, db *mongo.Database) ([]primitive.ObjectID, error) {
	cursor, err := db.Collection(schema.CollectionNameUser).Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"_id": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	userIDs := make([]primitive.ObjectID, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}
	sort.Slice(userIDs, func(i, j int) bool {
		return userIDs[i].Hex() < userIDs[j].Hex()
	})
	return userIDs, nil
}

func loadExistingChannelMemberships(ctx context.Context, db *mongo.Database) (map[primitive.ObjectID][]primitive.ObjectID, map[primitive.ObjectID][]primitive.ObjectID, error) {
	cursor, err := db.Collection(schema.CollectionNameChannelMembers).Find(ctx, bson.M{"status": schema.ChannelMemberStatusActive})
	if err != nil {
		return nil, nil, err
	}
	defer cursor.Close(ctx)

	var members []struct {
		ChannelID primitive.ObjectID `bson:"channel_id"`
		UserID    primitive.ObjectID `bson:"user_id"`
	}
	if err := cursor.All(ctx, &members); err != nil {
		return nil, nil, err
	}

	channelsByUser := make(map[primitive.ObjectID][]primitive.ObjectID)
	participantsByChannel := make(map[primitive.ObjectID][]primitive.ObjectID)
	seenUserChannel := make(map[string]bool, len(members))
	for _, member := range members {
		key := member.UserID.Hex() + ":" + member.ChannelID.Hex()
		if seenUserChannel[key] {
			continue
		}
		seenUserChannel[key] = true
		channelsByUser[member.UserID] = append(channelsByUser[member.UserID], member.ChannelID)
		participantsByChannel[member.ChannelID] = append(participantsByChannel[member.ChannelID], member.UserID)
	}
	return channelsByUser, participantsByChannel, nil
}

func loadExistingChannelSequences(ctx context.Context, db *mongo.Database) (map[primitive.ObjectID]int64, error) {
	cursor, err := db.Collection(schema.CollectionNameChannel).Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"_id": 1, "last_msg_seq": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var channels []struct {
		ID         primitive.ObjectID `bson:"_id"`
		LastMsgSeq int64              `bson:"last_msg_seq"`
	}
	if err := cursor.All(ctx, &channels); err != nil {
		return nil, err
	}

	seqs := make(map[primitive.ObjectID]int64, len(channels))
	for _, channel := range channels {
		seqs[channel.ID] = channel.LastMsgSeq
	}
	return seqs, nil
}

func buildAppendedMessages(rng *rand.Rand, userIDs []primitive.ObjectID, channelsByUser map[primitive.ObjectID][]primitive.ObjectID, channelSeqs map[primitive.ObjectID]int64, messagesPerUser int) ([]schema.Message, map[primitive.ObjectID]schema.Message) {
	now := time.Now()
	messagesByChannel := make(map[primitive.ObjectID][]schema.Message)
	for _, userID := range userIDs {
		channelIDs := channelsByUser[userID]
		if len(channelIDs) == 0 {
			continue
		}
		for i := 0; i < messagesPerUser; i++ {
			channelID := channelIDs[rng.Intn(len(channelIDs))]
			createdAt := randomRecentMessageTime(rng, now, messagesPerUser-i)
			updatedAt := createdAt.Add(time.Duration(rng.Intn(10*60)) * time.Second)
			if updatedAt.After(now) {
				updatedAt = now
			}
			content, msgType := messagePayload(rng, i+1)
			messagesByChannel[channelID] = append(messagesByChannel[channelID], schema.Message{
				ID:        primitive.NewObjectID(),
				ChannelID: channelID,
				FromID:    userID,
				Content:   content,
				MsgType:   msgType,
				Status:    messageStatus(rng),
				IsDelete:  false,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			})
		}
	}

	messages := make([]schema.Message, 0, len(userIDs)*messagesPerUser)
	lastByChannel := make(map[primitive.ObjectID]schema.Message, len(messagesByChannel))
	for channelID, channelMessages := range messagesByChannel {
		sort.Slice(channelMessages, func(i, j int) bool {
			return channelMessages[i].CreatedAt.Before(channelMessages[j].CreatedAt)
		})
		nextSeq := channelSeqs[channelID] + 1
		for i := range channelMessages {
			channelMessages[i].MsgSeq = nextSeq
			nextSeq++
			messages = append(messages, channelMessages[i])
			lastByChannel[channelID] = channelMessages[i]
		}
	}
	return messages, lastByChannel
}

func updateExistingChannelUnreads(ctx context.Context, db *mongo.Database, rng *rand.Rand, lastByChannel map[primitive.ObjectID]schema.Message, participantsByChannel map[primitive.ObjectID][]primitive.ObjectID) error {
	models := make([]mongo.WriteModel, 0)
	for channelID, lastMsg := range lastByChannel {
		for _, userID := range participantsByChannel[channelID] {
			update := bson.M{
				"$set": bson.M{
					"last_msg_id":   lastMsg.ID,
					"last_msg_time": lastMsg.CreatedAt,
					"is_active":     true,
					"version":       lastMsg.MsgSeq,
				},
				"$inc": bson.M{
					"unread": int64(rng.Intn(4)),
				},
				"$setOnInsert": bson.M{
					"_id":        primitive.NewObjectID(),
					"user_id":    userID,
					"channel_id": channelID,
				},
			}
			models = append(models, mongo.NewUpdateOneModel().
				SetFilter(bson.M{"user_id": userID, "channel_id": channelID}).
				SetUpdate(update).
				SetUpsert(true))
		}
	}
	if len(models) == 0 {
		return nil
	}
	_, err := db.Collection(schema.CollectionNameChannelUnread).BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
	return err
}

func buildChannelUnreads(rng *rand.Rand, groups []groupSeed, directs []directSeed, lastByChannel map[primitive.ObjectID]schema.Message) []schema.DbChannelUnread {
	unreads := make([]schema.DbChannelUnread, 0)
	addUnread := func(channelID primitive.ObjectID, participants []primitive.ObjectID) {
		lastMsg, ok := lastByChannel[channelID]
		if !ok {
			return
		}
		for _, userID := range participants {
			unreads = append(unreads, schema.DbChannelUnread{
				ID:          primitive.NewObjectID(),
				UserID:      userID,
				ChannelID:   channelID,
				LastMsgID:   lastMsg.ID,
				LastMsgTime: lastMsg.CreatedAt,
				IsActive:    true,
				Unread:      int64(rng.Intn(12)),
				Version:     lastMsg.MsgSeq,
			})
		}
	}
	for _, group := range groups {
		addUnread(group.channel.ID, group.participants)
	}
	for _, direct := range directs {
		addUnread(direct.channel.ID, direct.participants)
	}
	return unreads
}

func buildMessageOffsets(groups []groupSeed, directs []directSeed) []schema.MessageOffsets {
	offsets := make([]schema.MessageOffsets, 0)
	addOffset := func(channelID primitive.ObjectID, participants []primitive.ObjectID) {
		for _, userID := range participants {
			offsets = append(offsets, schema.MessageOffsets{
				ID:        primitive.NewObjectID(),
				UserID:    userID,
				ChannelID: channelID,
				Offset:    0,
				Version:   1,
				Sync:      true,
			})
		}
	}
	for _, group := range groups {
		addOffset(group.channel.ID, group.participants)
	}
	for _, direct := range directs {
		addOffset(direct.channel.ID, direct.participants)
	}
	return offsets
}

func buildMessageExtras(rng *rand.Rand, messages []schema.Message, participantsByChannel map[primitive.ObjectID][]primitive.ObjectID) []schema.MessageExtras {
	extras := make([]schema.MessageExtras, 0, len(messages))
	for _, msg := range messages {
		participants := participantsByChannel[msg.ChannelID]
		if len(participants) == 0 {
			continue
		}
		viewers := pickParticipants(rng, participants, min(3, len(participants)))
		for _, userID := range viewers {
			extras = append(extras, schema.MessageExtras{
				ID:        primitive.NewObjectID(),
				UserID:    userID,
				ChannelID: msg.ChannelID,
				MsgID:     msg.ID,
				Version:   msg.MsgSeq,
				Sync:      rng.Intn(10) > 1,
			})
		}
	}
	return extras
}

func buildMessageReactions(rng *rand.Rand, messages []schema.Message, participantsByChannel map[primitive.ObjectID][]primitive.ObjectID) []schema.MessageReaction {
	reactions := []schema.MessageReaction{}
	types := []string{"like", "love", "haha", "wow", "sad"}
	for _, msg := range messages {
		if rng.Intn(100) > 12 {
			continue
		}
		participants := participantsByChannel[msg.ChannelID]
		if len(participants) == 0 {
			continue
		}
		reactor := participants[rng.Intn(len(participants))]
		reactions = append(reactions, schema.MessageReaction{
			ID:             primitive.NewObjectID(),
			MsgID:          msg.ID,
			TypeOfReaction: types[rng.Intn(len(types))],
			CreatedBy:      reactor,
			CreatedAt:      msg.CreatedAt.Add(time.Duration(rng.Intn(20)) * time.Minute),
		})
	}
	return reactions
}

func channelParticipants(groups []groupSeed, directs []directSeed) map[primitive.ObjectID][]primitive.ObjectID {
	result := make(map[primitive.ObjectID][]primitive.ObjectID, len(groups)+len(directs))
	for _, group := range groups {
		result[group.channel.ID] = group.participants
	}
	for _, direct := range directs {
		result[direct.channel.ID] = direct.participants
	}
	return result
}

func insertMany(ctx context.Context, collection *mongo.Collection, docs []any, batchSize int) error {
	if len(docs) == 0 {
		return nil
	}
	for start := 0; start < len(docs); start += batchSize {
		end := start + batchSize
		if end > len(docs) {
			end = len(docs)
		}
		if _, err := collection.InsertMany(ctx, docs[start:end], options.InsertMany().SetOrdered(false)); err != nil {
			return err
		}
	}
	return nil
}

func docsFromChannelMembers(items []schema.DbChannelMember) []any {
	docs := make([]any, 0, len(items))
	for _, item := range items {
		docs = append(docs, item)
	}
	return docs
}

func docsFromMessages(items []schema.Message) []any {
	docs := make([]any, 0, len(items))
	for _, item := range items {
		docs = append(docs, item)
	}
	return docs
}

func docsFromUnreads(items []schema.DbChannelUnread) []any {
	docs := make([]any, 0, len(items))
	for _, item := range items {
		docs = append(docs, item)
	}
	return docs
}

func docsFromOffsets(items []schema.MessageOffsets) []any {
	docs := make([]any, 0, len(items))
	for _, item := range items {
		docs = append(docs, item)
	}
	return docs
}

func docsFromExtras(items []schema.MessageExtras) []any {
	docs := make([]any, 0, len(items))
	for _, item := range items {
		docs = append(docs, item)
	}
	return docs
}

func docsFromReactions(items []schema.MessageReaction) []any {
	docs := make([]any, 0, len(items))
	for _, item := range items {
		docs = append(docs, item)
	}
	return docs
}

func pickParticipants(rng *rand.Rand, userIDs []primitive.ObjectID, count int) []primitive.ObjectID {
	if count > len(userIDs) {
		count = len(userIDs)
	}
	indexes := rng.Perm(len(userIDs))[:count]
	participants := make([]primitive.ObjectID, 0, count)
	for _, index := range indexes {
		participants = append(participants, userIDs[index])
	}
	return participants
}

func sortedPair(a, b primitive.ObjectID) [2]primitive.ObjectID {
	pair := []primitive.ObjectID{a, b}
	sort.Slice(pair, func(i, j int) bool {
		return pair[i].Hex() < pair[j].Hex()
	})
	return [2]primitive.ObjectID{pair[0], pair[1]}
}

func hasUserBelowDegree(userIDs []primitive.ObjectID, degrees map[primitive.ObjectID]int, minDegree int) bool {
	for _, userID := range userIDs {
		if degrees[userID] < minDegree {
			return true
		}
	}
	return false
}

func randomBetween(rng *rand.Rand, minValue, maxValue int) int {
	if maxValue <= minValue {
		return minValue
	}
	return minValue + rng.Intn(maxValue-minValue+1)
}

func startOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func randomMessageTime(rng *rand.Rand, now time.Time, historyDays int) time.Time {
	dayOffset := rng.Intn(historyDays)
	dayStart := startOfDay(now).AddDate(0, 0, -dayOffset)
	minuteOfDay := randomBetween(rng, 8*60, 22*60+45)
	createdAt := dayStart.Add(time.Duration(minuteOfDay)*time.Minute + time.Duration(rng.Intn(60))*time.Second)
	if createdAt.After(now) {
		return now.Add(-time.Duration(rng.Intn(30)+1) * time.Second)
	}
	return createdAt
}

func randomRecentMessageTime(rng *rand.Rand, now time.Time, orderHint int) time.Time {
	if orderHint < 1 {
		orderHint = 1
	}
	return now.Add(-time.Duration(orderHint)*time.Minute + time.Duration(rng.Intn(45))*time.Second)
}

func avatarURL(index int) string {
	return fmt.Sprintf("https://api.dicebear.com/9.x/initials/svg?seed=go-app-%d", index)
}

func messagePayload(rng *rand.Rand, seq int) (string, schema.MessageType) {
	roll := rng.Intn(100)
	if roll < 8 {
		images := []string{
			"https://picsum.photos/seed/go-app-coffee/960/720",
			"https://picsum.photos/seed/go-app-meeting/960/720",
			"https://picsum.photos/seed/go-app-dinner/960/720",
			"https://picsum.photos/seed/go-app-desk/960/720",
			"https://picsum.photos/seed/go-app-trip/960/720",
		}
		return images[rng.Intn(len(images))], schema.MessageTypeImage
	}
	if roll < 13 {
		files := []string{
			"bao-cao-tuan-nay.pdf",
			"lich-hop-thang-7.xlsx",
			"hop-dong-ban-nhap.docx",
			"thiet-ke-man-hinh-chat.png",
			"danh-sach-cong-viec.xlsx",
		}
		return files[rng.Intn(len(files))], schema.MessageTypeFile
	}

	messages := []string{
		"Hom nay ban co ranh luc 3 gio khong?",
		"Minh vua xem lai roi, phan nay on do.",
		"Toi nay an gi? Minh dang nghi bun bo hoac com tam.",
		"Ban den noi chua? Neu ket xe thi cu bao minh.",
		"Mai minh nghi nua buoi sang, co gi nhan qua day nha.",
		"Deadline nay minh nghi day sang thu sau se hop ly hon.",
		"Cam on nha, cai nay giup minh tiet kiem kha nhieu thoi gian.",
		"Minh vua gui file, ban check giup minh phien ban moi nhat.",
		"Chuyen nay de minh hoi lai mot chut roi tra loi ban sau.",
		"Cuoi tuan nay nha minh co viec nen minh xin phep vang.",
		"Ok, minh note lai roi. Chieu minh update trang thai tiep.",
		"Nghe hop ly do, nhung minh muon test them tren mobile.",
		"Ban nho uong nuoc voi nghi mat chut nha, lam tu sang toi gio roi.",
		"Minh se qua don ban luc 7 gio, neu co thay doi thi nhan minh.",
		"Phan login da chay on, con flow refresh token minh dang xem tiep.",
		"Khach vua phan hoi, ho muon text ngan gon va de hieu hon.",
		"Minh thay mau nay dep hon, nhin sach va de doc.",
		"Nhan duoc roi nha, de minh doc ky roi comment tung muc.",
		"Hom qua ve muon qua nen minh chua kip tra loi.",
		"Neu can gap thi goi minh, con khong thi de toi ve minh xem.",
		"Minh vua dat xe roi, khoang 10 phut nua toi noi.",
		"Cho minh xin lai link meeting voi, lich cua minh bi mat invite.",
		"Phan nay nen tach thanh task rieng de de review hon.",
		"Minh dong y, nhung can them log de debug luc co loi.",
		"Toi nay minh nau com o nha, ban ghe qua an chung khong?",
		"Sang mai minh se day som xu ly not viec nay.",
		"Dung quen backup database truoc khi reset moi truong test nha.",
		"Tin nay hay do, gui them cho minh bai goc duoc khong?",
		"Minh vua sua xong bug nho, ban pull lai roi test thu.",
		"Khong sao dau, cu tu tu. Viec nay khong can gap qua.",
		"Ban thay avatar nay co bi mo qua khong?",
		"Minh can them du lieu mau de man hinh chat nhin that hon.",
		"Chot vay nha, 9 gio sang mai minh bat dau.",
		"Da thanh toan xong roi, minh gui bill sau.",
		"Me vua goi, toi nay minh ve nha an com.",
		"Minh dang tren duong, co le tre khoang 5 phut.",
		"Nho mang laptop voi sac nha, phong hop khong co may du phong.",
		"Ban review giup minh phan wording trong popup nay.",
		"Minh vua cap nhat lai UI, nhin bot bi rong hon truoc.",
		"Neu data nhieu qua thi minh se them pagination cho nhe.",
	}

	return fmt.Sprintf("%s (%02d)", messages[rng.Intn(len(messages))], seq), schema.MessageTypeText
}

func messageStatus(rng *rand.Rand) schema.MessageStatus {
	switch rng.Intn(100) {
	case 0, 1, 2, 3, 4:
		return schema.MessageStatusSent
	case 5, 6, 7, 8, 9, 10, 11, 12:
		return schema.MessageStatusRead
	default:
		return schema.MessageStatusDelivered
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
