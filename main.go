package main

import (
	"fmt"
	"log"
	"os"
	"songBot/src"
	"songBot/src/config"
	"strconv"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
)

var (
	startTimeStamp = time.Now()
)

func main() {
	checkEnvVars(
		map[string]string{
			"TOKEN":    config.Token,
			"API_KEY":  config.ApiKey,
			"API_HASH": config.ApiHash,
			"API_ID":   config.ApiId,
			"API_URL":  config.ApiUrl,
		},
	)

	createDir("downloads")

	client, ok := buildAndStart(0, config.Token)
	if !ok {
		log.Fatal("❌ [Startup] Bot client initialization failed")
	}

	client.Idle()
	log.Println("🛑 [Shutdown] Bot stopped.")
}

// checkEnvVars validates required environment variables
func checkEnvVars(vars map[string]string) {
	for k, v := range vars {
		if v == "" {
			log.Fatalf("❌ Missing required environment variable: %s", k)
		}
	}
}

// createDir ensures the directory exists or creates it
func createDir(dir string) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Fatalf("❌ Failed to create directory %s: %v", dir, err)
	}
}

// buildAndStart initializes and logs into the bot client
func buildAndStart(index int, token string) (*tg.Client, bool) {
	apiId, err := strconv.Atoi(config.ApiId)
	if err != nil {
		log.Printf("[Client %d] ❌ Invalid API_ID: %v", index, err)
		return nil, false
	}

	client, err := tg.NewClient(tg.ClientConfig{
		AppID:   int32(apiId),
		AppHash: config.ApiHash,
		//FloodHandler: handleFlood,
		SessionName: fmt.Sprintf("bot_%d", index),
	})
	if err != nil {
		log.Printf("[Client %d] ❌ Client creation failed: %v", index, err)
		return nil, false
	}

	if _, err := client.Conn(); err != nil {
		log.Printf("[Client %d] ❌ Connection error: %v", index, err)
		return nil, false
	}

	if err := client.LoginBot(token); err != nil {
		log.Printf("[Client %d] ❌ Bot login failed: %v", index, err)
		return nil, false
	}

	me, err := client.GetMe()
	if err != nil {
		log.Printf("[Client %d] ❌ Failed to fetch bot info: %v", index, err)
		return nil, false
	}

	uptime := time.Since(startTimeStamp).Round(time.Millisecond)
	log.Printf("✅ [Client %d] Logged in as @%s | Startup time: %s", index, me.Username, uptime)

	src.InitFunc(client)
	return client, true
}

// handleFlood delays on flood wait errors
func handleFlood(err error) bool {
	if wait := tg.GetFloodWait(err); wait > 0 {
		log.Printf("⚠️ Flood wait detected: sleeping for %ds", wait)
		time.Sleep(time.Duration(wait) * time.Second)
		return true
	}
	return false
}
