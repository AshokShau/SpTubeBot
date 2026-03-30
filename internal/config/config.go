package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	ApiId   int32
	ApiHash string
	ApiKey  string
	ApiUrl  string
	Token    string
	OwnerId  int64
	MongoUri string
}

var ErrMissingEnv = errors.New("missing required environment variable")

func Load() (*Config, error) {
	cfg := &Config{}
	apiId, err := requireEnv("API_ID")
	if err != nil {
		return nil, err
	}

	id, err := strconv.ParseInt(apiId, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("API_ID must be a valid integer: %w", err)
	}

	cfg.ApiId = int32(id)

	cfg.ApiHash, err = requireEnv("API_HASH")
	if err != nil {
		return nil, err
	}

	cfg.ApiKey, err = requireEnv("API_KEY")
	if err != nil {
		return nil, err
	}

	cfg.Token, err = requireEnv("TOKEN")
	if err != nil {
		return nil, err
	}

	cfg.ApiUrl = os.Getenv("API_URL")
	if cfg.ApiUrl == "" {
		cfg.ApiUrl = "https://api.fallenapi.fun"
	}

	ownerId := os.Getenv("OWNER_ID")
	if ownerId == "" {
		ownerId = "5938660179"
	}

	oid, err := strconv.ParseInt(ownerId, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("OWNER_ID must be a valid integer: %w", err)
	}

	cfg.OwnerId = oid

	cfg.MongoUri = os.Getenv("MONGO_URL")
	if cfg.MongoUri == "" {
		return nil, fmt.Errorf("%w: MONGO_URL", ErrMissingEnv)
	}

	return cfg, nil
}

func requireEnv(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("%w: %s", ErrMissingEnv, key)
	}

	return val, nil
}
