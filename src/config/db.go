package config

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	databaseName   = "songBot"
	collectionName = "users"
	timeout        = 10 * time.Second
)

var (
	client     *mongo.Client
	collection *mongo.Collection
	ctx        = context.Background()
)

// User schema
type User struct {
	ID       int64  `bson:"_id"`
	BotToken string `bson:"bot_token,omitempty"`
}

func init() {
	if MongoUrl == "" {
		log.Fatal("MongoUrl is not set")
	}

	c, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var err error
	client, err = mongo.Connect(c, options.Client().ApplyURI(MongoUrl))
	if err != nil {
		log.Fatalf("MongoDB connection error: %v", err)
	}

	if err = client.Ping(c, nil); err != nil {
		log.Fatalf("MongoDB ping failed: %v", err)
	}

	collection = client.Database(databaseName).Collection(collectionName)

	// Create index on bot_token (optional but useful for lookups)
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"bot_token": 1},
		Options: options.Index().SetUnique(false),
	})
	if err != nil {
		log.Printf("⚠️ Failed to create index on bot_token: %v", err)
	}
}

// SaveUser inserts user ID if not already present
func SaveUser(userID int64) error {
	_, err := collection.UpdateByID(
		ctx,
		userID,
		bson.M{"$setOnInsert": bson.M{"bot_token": nil}},
		options.Update().SetUpsert(true),
	)
	return err
}

// AddBotToken stores the bot token if not already added by any user
func AddBotToken(userID int64, token string) error {
	if token == "" {
		return errors.New("token is empty")
	}

	exists, err := TokenExists(token)
	if err != nil {
		return err
	}
	if exists {
		log.Printf("⚠️ Token already exists in database. Skipping insert.")
		return nil
	}

	_, err = collection.UpdateByID(
		ctx,
		userID,
		bson.M{"$set": bson.M{"bot_token": token}},
		options.Update().SetUpsert(true),
	)
	return err
}

// TokenExists checks whether a token is already stored
func TokenExists(token string) (bool, error) {
	filter := bson.M{"bot_token": token}
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetAllUserIDs returns list of all user IDs
func GetAllUserIDs() ([]int64, error) {
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}
	return ids, nil
}

// GetAllBotTokens returns list of all non-empty bot tokens
func GetAllBotTokens() ([]string, error) {
	filter := bson.M{"bot_token": bson.M{"$ne": nil}}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	tokens := make([]string, 0, len(users))
	for _, u := range users {
		if u.BotToken != "" {
			tokens = append(tokens, u.BotToken)
		}
	}
	return tokens, nil
}

// GetBotTokenByUserID fetches the bot token associated with a specific user ID
func GetBotTokenByUserID(userID int64) (string, error) {
	var user User
	err := collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", nil
		}
		return "", err
	}
	return user.BotToken, nil
}

// RemoveBotToken unsets the bot_token field for the user that owns the given token
func RemoveBotToken(token string) error {
	if token == "" {
		return errors.New("token is empty")
	}

	filter := bson.M{"bot_token": token}
	update := bson.M{"$unset": bson.M{"bot_token": ""}}

	res, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("no user found with this token")
	}
	return nil
}
