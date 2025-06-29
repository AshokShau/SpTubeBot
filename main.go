package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"songBot/src"
	"songBot/src/config"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
)

var (
	startTimeStamp = time.Now().Unix()
	restartClient  = &http.Client{Timeout: 10 * time.Second}
)

func main() {
	if config.Token == "" || config.ApiKey == "" || config.ApiUrl == "" || config.MongoUrl == "" {
		log.Fatal("Missing environment variables. Please set TOKEN, API_KEY, API_URL and MONGO_URL.")
	}

	if err := os.Mkdir("downloads", os.ModePerm); err != nil && !os.IsExist(err) {
		log.Fatalf("Failed to create downloads directory: %v", err)
	}

	Tokens, err := config.GetAllBotTokens()
	if err == nil && len(Tokens) > 0 {
		log.Printf("Starting %d clients.", len(Tokens))
		for i := 1; i <= len(Tokens); i++ {
			go buildAndStart(i, Tokens[i-1])
		}
	} else if err != nil {
		log.Printf("Failed to retrieve bot tokens: %v", err)
	} else {
		log.Println("No bot tokens found.")
	}

	// Start main client
	client, ok := buildAndStart(0, config.Token)
	if !ok {
		log.Fatalf("[Client] Startup failed")
	}

	go autoRestart(24 * time.Hour)
	client.Idle()
	log.Printf("[Client 0] Bot stopped.")
}

func buildAndStart(index int, token string) (*tg.Client, bool) {
	clientConfig := tg.ClientConfig{
		AppID:         8,
		AppHash:       "7245de8e747a0d6fbe11f7cc14fcc0bb",
		MemorySession: true,
		FloodHandler:  handleFlood,
		SessionName:   fmt.Sprintf("bot_%d", index),
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
		_ = config.RemoveBotToken(token)
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

func autoRestart(interval time.Duration) {
	if config.CoolifyToken == "" {
		log.Println("Coolify token not set; autoRestart disabled.")
		return
	}

	go func() {
		for {
			time.Sleep(interval)
			req, err := http.NewRequest("GET",
				"https://app.ashok.sbs/api/v1/applications/lkkgog40occ0c8soo8gwcokk/restart", nil)
			if err != nil {
				log.Printf("[Restart] ❌ Request error: %v", err)
				continue
			}
			req.Header.Set("Authorization", "Bearer "+config.CoolifyToken)

			resp, err := restartClient.Do(req)
			if err != nil {
				log.Printf("[Restart] ❌ Request failed: %v", err)
				continue
			}
			_ = resp.Body.Close()
			log.Printf("[Restart] ✅ Status: %s", resp.Status)
		}
	}()
}

func handleFlood(err error) bool {
	if wait := tg.GetFloodWait(err); wait > 0 {
		time.Sleep(time.Duration(wait) * time.Second)
		return true
	}
	return false
}
