package main

//go:generate go run github.com/AshokShau/gotdbot/scripts/tools@latest

import (
	"log"
	"noinoi/internal/commands"
	"noinoi/internal/config"
	"noinoi/internal/database"
	"noinoi/internal/httpx"
	"strings"

	"github.com/AshokShau/gotdbot"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	err = database.Init(cfg.MongoUri)
	if err != nil {
		log.Fatal(err)
	}

	httpx.Init(cfg.ApiKey, cfg.ApiUrl)

	manager := gotdbot.NewClientManager("./libtdjson.so.1.8.62")
	dispatcher := gotdbot.NewDispatcher(nil)
	commands.LoadCmd(dispatcher, manager, cfg)

	type botToRegister struct {
		token   string
		dbDir   string
		ownerID int64
	}

	var botsToRegister []botToRegister
	botsToRegister = append(botsToRegister, botToRegister{
		token:   cfg.Token,
		dbDir:   "db_main",
		ownerID: cfg.OwnerId,
	})

	dbBots, err := database.GetAllBots()
	if err == nil {
		for _, b := range dbBots {
			botsToRegister = append(botsToRegister, botToRegister{
				token:   b.BotToken,
				dbDir:   "db_" + strings.Split(b.BotToken, ":")[0],
				ownerID: b.UserId,
			})
		}
	} else {
		log.Printf("Failed to get all bots from database: %v", err)
	}

	for _, b := range botsToRegister {
		con := gotdbot.DefaultClientConfig()
		con.Dispatcher = dispatcher
		con.DatabaseDirectory = b.dbDir
		client, err := manager.RegisterClient(cfg.ApiId, cfg.ApiHash, b.token, con)
		if err != nil {
			log.Printf("Failed to register client for token %s: %v", b.token, err)
			_ = database.DeleteBot(b.ownerID)
			continue
		}

		me, err := client.GetMe()
		if err != nil {
			log.Printf("Failed to get me for token %s: %v", b.token, err)
			continue
		}

		username := ""
		if me.Usernames != nil && len(me.Usernames.ActiveUsernames) > 0 {
			username = me.Usernames.ActiveUsernames[0]
		}
		client.Logger.Info("Logged in", "username", username, "id", me.Id)
	}

	manager.Idle()
}
