package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"songBot/src"
	"songBot/src/config"
	"strconv"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
)

var (
	startTimeStamp = time.Now().Unix()
	restartClient  = &http.Client{Timeout: 10 * time.Second}
)

func main() {
	if config.Token == "" || config.ApiKey == "" || config.ApiUrl == "" || config.ApiHash == "" || config.ApiId == "" {
		log.Fatal("Missing environment variables. Please set TOKEN, API_KEY, API_HASH, API_ID and API_URL")
	}

	if err := os.Mkdir("downloads", os.ModePerm); err != nil && !os.IsExist(err) {
		log.Fatalf("Failed to create downloads directory: %v", err)
	}

	client, ok := buildAndStart(0, config.Token)
	if !ok {
		log.Fatalf("[Client] Startup failed")
	}

	client.Idle()
	log.Printf("[Client] Bot stopped.")
}

func buildAndStart(index int, token string) (*tg.Client, bool) {
	apiId, err := strconv.Atoi(config.ApiId)
	if err != nil {
		log.Printf("[Client %d] ❌ Failed to parse API ID: %v", index, err)
		return nil, false
	}

	clientConfig := tg.ClientConfig{
		AppID:        int32(apiId),
		AppHash:      config.ApiHash,
		FloodHandler: handleFlood,
		SessionName:  fmt.Sprintf("bot_%d", index),
	}

	client, err := tg.NewClient(clientConfig)
	if err != nil {
		log.Printf("[Client %d] ❌ Failed to create client: %v", index, err)
		return nil, false
	}

	if _, err = client.Conn(); err != nil {
		log.Printf("[Client %d] ❌ Connection error: %v", index, err)
		return nil, false
	}

	if err = client.LoginBot(token); err != nil {
		log.Printf("[Client %d] ❌ Bot login failed: %v", index, err)
		return nil, false
	}

	me, err := client.GetMe()
	if err != nil {
		log.Printf("[Client %d] ❌ Failed to get bot info: %v", index, err)
		return nil, false
	}

	uptime := time.Since(time.Unix(startTimeStamp, 0)).String()
	client.Logger.Info(fmt.Sprintf("✅ Client %d: @%s (Startup in %s)", index, me.Username, uptime))
	src.InitFunc(client)
	return client, true
}

func handleFlood(err error) bool {
	if wait := tg.GetFloodWait(err); wait > 0 {
		time.Sleep(time.Duration(wait) * time.Second)
		return true
	}
	return false
}
