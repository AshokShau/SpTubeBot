package database

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type BotInfo struct {
	BotId     int64     `bson:"_id"` // bot_id
	OwnerId   int64     `bson:"owner_id"`
	BotToken  string    `bson:"bot_token"`
	CreatedAt time.Time `bson:"created_at"`
}

type BotStats struct {
	BotId            int64   `bson:"_id"`
	Users            []int64 `bson:"users"`
	Chats            []int64 `bson:"chats"`
	SuccessDownloads int64   `bson:"success_downloads"`
	FailedDownloads  int64   `bson:"failed_downloads"`
}

var (
	client              *mongo.Client
	collection          *mongo.Collection
	statsCollection     *mongo.Collection
	blacklistCollection *mongo.Collection
	botToOwner          = make(map[int64]int64)
	ownersMu            sync.RWMutex

	trackedUsers = make(map[int64]map[int64]bool)
	trackedChats = make(map[int64]map[int64]bool)
	trackedMu    sync.RWMutex

	blacklistedIDs = make(map[int64]map[int64]bool)
	blacklistMu    sync.RWMutex
)

func Init(uri string) error {
	var err error
	client, err = mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	db := client.Database("noinoi")
	collection = db.Collection("bots")
	statsCollection = db.Collection("stats")
	blacklistCollection = db.Collection("blacklist")

	go preloadStats()
	go preloadBlacklist()

	return nil
}

func preloadStats() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := statsCollection.Find(ctx, bson.M{})
	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	trackedMu.Lock()
	defer trackedMu.Unlock()

	for cursor.Next(ctx) {
		var s BotStats
		if err = cursor.Decode(&s); err != nil {
			continue
		}
		if trackedUsers[s.BotId] == nil {
			trackedUsers[s.BotId] = make(map[int64]bool)
		}
		for _, u := range s.Users {
			trackedUsers[s.BotId][u] = true
		}
		if trackedChats[s.BotId] == nil {
			trackedChats[s.BotId] = make(map[int64]bool)
		}
		for _, c := range s.Chats {
			trackedChats[s.BotId][c] = true
		}
	}
}

func ParseBotId(token string) int64 {
	split := strings.Split(token, ":")
	if len(split) < 1 {
		return 0
	}
	id, _ := strconv.ParseInt(split[0], 10, 64)
	return id
}

func SaveBot(bot BotInfo) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Replace().SetUpsert(true)
	filter := bson.M{"_id": bot.BotId}
	_, err := collection.ReplaceOne(ctx, filter, bot, opts)
	if err == nil {
		SetOwner(bot.BotId, bot.OwnerId)
	}
	return err
}

func DeleteBot(botId int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": botId}
	_, err := collection.DeleteOne(ctx, filter)
	if err == nil {
		ownersMu.Lock()
		delete(botToOwner, botId)
		ownersMu.Unlock()
	}
	return err
}

func GetBotsByOwner(ownerId int64) ([]BotInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"owner_id": ownerId})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bots []BotInfo
	if err = cursor.All(ctx, &bots); err != nil {
		return nil, err
	}

	return bots, nil
}

func GetAllBots() ([]BotInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bots []BotInfo
	if err = cursor.All(ctx, &bots); err != nil {
		return nil, err
	}

	for _, b := range bots {
		SetOwner(b.BotId, b.OwnerId)
	}

	return bots, nil
}

func SetOwner(botId, ownerId int64) {
	ownersMu.Lock()
	defer ownersMu.Unlock()
	botToOwner[botId] = ownerId
}

func GetOwner(botId int64) (int64, bool) {
	ownersMu.RLock()
	defer ownersMu.RUnlock()
	ownerId, ok := botToOwner[botId]
	return ownerId, ok
}

func IncrementDownloads(botId int64, success bool) {
	if botId == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	field := "failed_downloads"
	if success {
		field = "success_downloads"
	}

	filter := bson.M{"_id": botId}
	update := bson.M{"$inc": bson.M{field: 1}}
	opts := options.UpdateOne().SetUpsert(true)

	_, _ = statsCollection.UpdateOne(ctx, filter, update, opts)
}

func AddUserOrChat(botId int64, id int64, isPrivate bool) {
	if id == 0 {
		return
	}

	trackedMu.Lock()
	if isPrivate {
		if trackedUsers[botId] == nil {
			trackedUsers[botId] = make(map[int64]bool)
		}
		if trackedUsers[botId][id] {
			trackedMu.Unlock()
			return
		}
		trackedUsers[botId][id] = true
	} else {
		if trackedChats[botId] == nil {
			trackedChats[botId] = make(map[int64]bool)
		}
		if trackedChats[botId][id] {
			trackedMu.Unlock()
			return
		}
		trackedChats[botId][id] = true
	}
	trackedMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	field := "chats"
	if isPrivate {
		field = "users"
	}

	filter := bson.M{"_id": botId}
	update := bson.M{"$addToSet": bson.M{field: id}}
	opts := options.UpdateOne().SetUpsert(true)

	_, _ = statsCollection.UpdateOne(ctx, filter, update, opts)
}

func GetBotStats(botId int64) (BotStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var stats BotStats
	err := statsCollection.FindOne(ctx, bson.M{"_id": botId}).Decode(&stats)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return BotStats{BotId: botId}, nil
		}
		return stats, err
	}
	return stats, nil
}

func GetUsers(botId int64) []int64 {
	trackedMu.RLock()
	defer trackedMu.RUnlock()

	usersMap, ok := trackedUsers[botId]
	if !ok {
		return nil
	}

	users := make([]int64, 0, len(usersMap))
	for id := range usersMap {
		users = append(users, id)
	}
	return users
}

func GetChats(botId int64) []int64 {
	trackedMu.RLock()
	defer trackedMu.RUnlock()

	chatsMap, ok := trackedChats[botId]
	if !ok {
		return nil
	}

	chats := make([]int64, 0, len(chatsMap))
	for id := range chatsMap {
		chats = append(chats, id)
	}
	return chats
}

func GetGlobalStats() (totalBots int64, totalUsers int, totalChats int, totalSuccess int64, totalFailed int64, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	totalBots, err = collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return
	}

	cursor, err := statsCollection.Find(ctx, bson.M{})
	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	allUsers := make(map[int64]bool)
	allChats := make(map[int64]bool)

	for cursor.Next(ctx) {
		var s BotStats
		if err = cursor.Decode(&s); err != nil {
			return
		}
		for _, u := range s.Users {
			allUsers[u] = true
		}
		for _, c := range s.Chats {
			allChats[c] = true
		}
		totalSuccess += s.SuccessDownloads
		totalFailed += s.FailedDownloads
	}

	totalUsers = len(allUsers)
	totalChats = len(allChats)
	return
}

func preloadBlacklist() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := blacklistCollection.Find(ctx, bson.M{})
	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	blacklistMu.Lock()
	defer blacklistMu.Unlock()

	for cursor.Next(ctx) {
		var b struct {
			BotID    int64 `bson:"bot_id"`
			EntityID int64 `bson:"entity_id"`
		}
		if err = cursor.Decode(&b); err != nil {
			continue
		}
		if blacklistedIDs[b.BotID] == nil {
			blacklistedIDs[b.BotID] = make(map[int64]bool)
		}
		blacklistedIDs[b.BotID][b.EntityID] = true
	}
}

func BlacklistEntity(botID, entityID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"bot_id": botID, "entity_id": entityID}
	update := bson.M{"$set": bson.M{"bot_id": botID, "entity_id": entityID}}
	opts := options.UpdateOne().SetUpsert(true)

	_, err := blacklistCollection.UpdateOne(ctx, filter, update, opts)
	if err == nil {
		blacklistMu.Lock()
		if blacklistedIDs[botID] == nil {
			blacklistedIDs[botID] = make(map[int64]bool)
		}
		blacklistedIDs[botID][entityID] = true
		blacklistMu.Unlock()
	}
	return err
}

func WhitelistEntity(botID, entityID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"bot_id": botID, "entity_id": entityID}
	_, err := blacklistCollection.DeleteOne(ctx, filter)
	if err == nil {
		blacklistMu.Lock()
		if blacklistedIDs[botID] != nil {
			delete(blacklistedIDs[botID], entityID)
		}
		blacklistMu.Unlock()
	}
	return err
}

func IsBlacklisted(botID, entityID int64) bool {
	blacklistMu.RLock()
	defer blacklistMu.RUnlock()
	if blacklistedIDs[botID] == nil {
		return false
	}
	return blacklistedIDs[botID][entityID]
}
