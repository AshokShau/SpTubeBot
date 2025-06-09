package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"songBot/src/config"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
)

var (
	startTimeStamp = time.Now().Unix()
	restartClient  = &http.Client{
		Timeout: 10 * time.Second,
	}
)

func autoRestart(interval time.Duration) {
	if config.CoolifyToken == "" {
		log.Println("Coolify token not set; autoRestart disabled.")
		return
	}

	go func() {
		for {
			time.Sleep(interval)
			restartURL := "https://app.ashok.sbs/api/v1/applications/lkkgog40occ0c8soo8gwcokk/restart"
			req, err := http.NewRequest("GET", restartURL, nil)
			if err != nil {
				log.Printf("[Restart] Failed to create request: %v", err)
				continue
			}

			req.Header.Set("Authorization", "Bearer "+config.CoolifyToken)

			resp, err := restartClient.Do(req)
			if err != nil {
				log.Printf("[Restart] Failed to make request: %v", err)
				continue
			}
			_ = resp.Body.Close()
			log.Printf("[Restart] Status: %s", resp.Status)
		}
	}()
}

func main() {
	if config.Token == "" || config.ApiKey == "" || config.ApiUrl == "" {
		log.Fatal("Missing environment variables. Please set TOKEN, API_KEY, and API_URL.")
	}

	// https://pastebin.com/0mWJ7MWQ
	clientConfig := tg.ClientConfig{
		AppID:        8,
		AppHash:      "7245de8e747a0d6fbe11f7cc14fcc0bb",
		Session:      "session.dat",
		FloodHandler: handleFlood,
	}

	client, err := tg.NewClient(clientConfig)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	if _, err = client.Conn(); err != nil {
		log.Fatalf("Failed to connect client: %v", err)
	}

	if err = client.LoginBot(config.Token); err != nil {
		log.Fatalf("Bot login failed: %v", err)
	}

	if err = os.Mkdir("downloads", os.ModePerm); err != nil && !os.IsExist(err) {
		log.Fatalf("Failed to create downloads directory: %v", err)
	}

	initFunc(client)

	me, err := client.GetMe()
	if err != nil {
		log.Fatalf("Failed to get bot information: %v", err)
	}
	go autoRestart(6 * time.Hour)
	uptime := time.Since(time.Unix(startTimeStamp, 0)).String()
	client.Logger.Info(fmt.Sprintf("Authenticated as -> @%s, taken: %s.", me.Username, uptime))
	client.Logger.Info("GoGram version: " + tg.Version)
	client.Idle()
	log.Println("Bot stopped.")
}

func handleFlood(err error) bool {
	wait := tg.GetFloodWait(err)
	if wait > 0 {
		time.Sleep(time.Duration(wait) * time.Second)
		return true
	}
	return false
}
