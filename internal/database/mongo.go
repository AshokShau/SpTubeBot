package database

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type BotInfo struct {
	UserId    int64     `bson:"_id"`
	BotToken  string    `bson:"bot_token"`
	CreatedAt time.Time `bson:"created_at"`
}

var (
	client     *mongo.Client
	collection *mongo.Collection
	botToOwner = make(map[int64]int64)
	ownerToBot = make(map[int64]int64)
	ownersMu   sync.RWMutex
)

func Init(uri string) error {
	var err error
	client, err = mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	collection = client.Database("noinoi").Collection("bots")
	return nil
}

func parseBotId(token string) int64 {
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
	filter := bson.M{"_id": bot.UserId}
	_, err := collection.ReplaceOne(ctx, filter, bot, opts)
	if err == nil {
		botId := parseBotId(bot.BotToken)
		if botId != 0 {
			setOwner(botId, bot.UserId)
		}
	}
	return err
}

func DeleteBot(userId int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": userId}
	_, err := collection.DeleteOne(ctx, filter)
	if err == nil {
		ownersMu.Lock()
		if botId, ok := ownerToBot[userId]; ok {
			delete(botToOwner, botId)
			delete(ownerToBot, userId)
		}
		ownersMu.Unlock()
	}
	return err
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
		botId := parseBotId(b.BotToken)
		if botId != 0 {
			setOwner(botId, b.UserId)
		}
	}

	return bots, nil
}

func setOwner(botId, ownerId int64) {
	ownersMu.Lock()
	defer ownersMu.Unlock()
	botToOwner[botId] = ownerId
	ownerToBot[ownerId] = botId
}

func GetOwner(botId int64) (int64, bool) {
	ownersMu.RLock()
	defer ownersMu.RUnlock()
	ownerId, ok := botToOwner[botId]
	return ownerId, ok
}
