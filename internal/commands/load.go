package commands

import (
	"fmt"
	"noinoi/internal/config"
	"noinoi/internal/database"
	"noinoi/internal/httpx"
	"strconv"
	"strings"
	"time"

	"github.com/AshokShau/gotdbot"
)

var (
	startTime    = time.Now()
	globalConfig *config.Config
	manager      *gotdbot.ClientManager
)

func SetupHandlers(c *gotdbot.Client, m *gotdbot.ClientManager, cfg *config.Config) {
	globalConfig = cfg
	manager = m

	c.OnCommand("ping", pingHandler)
	c.OnCommand("start", startHandler)
	c.OnCommand("help", startHandler)
	c.OnCommand("yt", ytCommandHandler)
	c.OnCommand("math", mathHandler)
	c.OnCommand("stop", stopHandler)
	c.OnCommand("stats", statsHandler)
	c.OnCommand("catbox", catboxHandler)
	c.OnCommand("tgm", catboxHandler)
	c.OnCommand("litterbox", litterboxHandler)
	c.OnCommand("block", blockHandler)
	c.OnCommand("unblock", unblockHandler)
	c.OnCommand("broadcast", broadcastHandler)
	c.OnCommand("stopbroadcast", cancelBroadcastHandler)

	c.OnUpdateNewInlineQuery(handleInlineQuery, nil)
	c.OnUpdateNewInlineCallbackQuery(handleInlineCallbackQuery, nil)
	c.OnUpdateNewGuestQuery(handleGuestQuery, nil)

	c.OnMessage(func(c *gotdbot.Client, msg *gotdbot.Message) error {
		senderID := msg.SenderID()
		if senderID != globalConfig.OwnerId && (database.IsBlacklisted(c.Me.Id, senderID) || database.IsBlacklisted(c.Me.Id, msg.ChatId)) {
			return gotdbot.EndGroups
		}

		text := msg.GetText()

		if httpx.YouTubeShortsPattern.MatchString(text) || httpx.YouTubePattern.MatchString(text) || httpx.YouTubePostPattern.MatchString(text) {
			return youtubeHandler(c, msg)
		}

		for _, pattern := range httpx.SnapPatterns {
			if pattern.MatchString(text) {
				return snapHandler(c, msg)
			}
		}

		for _, pattern := range httpx.MusicPatterns {
			if pattern.MatchString(text) {
				return musicHandler(c, msg)
			}
		}

		return gotdbot.EndGroups
	}, func(msg *gotdbot.Message) bool {
		if msg == nil || msg.IsCommand() {
			return false
		}

		text := msg.GetText()
		if text == "" {
			return false
		}

		if httpx.YouTubeShortsPattern.MatchString(text) || httpx.YouTubePattern.MatchString(text) || httpx.YouTubePostPattern.MatchString(text) {
			return true
		}

		for _, pattern := range httpx.SnapPatterns {
			if pattern.MatchString(text) {
				return true
			}
		}

		for _, pattern := range httpx.MusicPatterns {
			if pattern.MatchString(text) {
				return true
			}
		}

		return false
	})

	c.OnUpdateNewCallbackQuery(handleCloneCreate, func(u *gotdbot.UpdateNewCallbackQuery) bool {
		return u.DataString() == "clone_create"
	})

	c.OnUpdateNewCallbackQuery(handleMyBots, func(u *gotdbot.UpdateNewCallbackQuery) bool {
		return u.DataString() == "clone_mybots"
	})

	c.OnUpdateNewCallbackQuery(handleBotManage, func(u *gotdbot.UpdateNewCallbackQuery) bool {
		return strings.HasPrefix(u.DataString(), "bot_")
	})

	c.OnUpdateNewCallbackQuery(handleBotRevoke, func(u *gotdbot.UpdateNewCallbackQuery) bool {
		return strings.HasPrefix(u.DataString(), "revoke_")
	})

	c.OnUpdateNewCallbackQuery(handleBotDelete, func(u *gotdbot.UpdateNewCallbackQuery) bool {
		return strings.HasPrefix(u.DataString(), "delete_")
	})

	c.OnUpdateNewCallbackQuery(handleCloneBack, func(u *gotdbot.UpdateNewCallbackQuery) bool {
		data := u.CallbackData()
		return string(data) == "clone_back"
	})

	c.OnUpdateManagedBot(func(c *gotdbot.Client, u *gotdbot.UpdateManagedBot) error {
		log := c.Logger.With("bot_id", u.BotUserId, "owner_id", u.UserId)
		log.Info("Managed bot updated")

		botToken, err := c.GetManagedBotToken(u.BotUserId, nil)
		if err != nil {
			log.Error("Failed to get token for managed bot", "error", err)
			return nil
		}

		if err = database.SaveBot(database.BotInfo{
			BotId:     u.BotUserId,
			OwnerId:   u.UserId,
			BotToken:  botToken.Text,
			CreatedAt: time.Now(),
		}); err != nil {
			log.Error("Failed to save managed bot to DB", "error", err)
			return nil
		}

		if isClientRunning(m, u.BotUserId) {
			log.Info("Managed bot is already running, skipping start")
			return nil
		}

		clientConfig := gotdbot.DefaultClientConfig()
		clientConfig.DatabaseDirectory = "db_" + strconv.FormatInt(u.BotUserId, 10)

		newBot, err := m.RegisterClient(cfg.ApiId, cfg.ApiHash, botToken.Text, clientConfig)
		if err != nil {
			log.Error("Failed to start managed clone bot", "error", err)
			_, _ = c.SendTextMessage(u.UserId, "Your bot was created, but I failed to start the clone.", nil)
			return nil
		}

		SetupHandlers(newBot, m, cfg)

		username := newBot.Me.Usernames.EditableUsername
		_, _ = c.SendTextMessage(u.UserId, fmt.Sprintf("Your clone bot %s (@%s) is now up and running! 🎉", newBot.Me.FirstName, username), nil)
		return nil
	}, nil)

	c.AddUpdateNewMessageHandlerGroup(func(c *gotdbot.Client, m *gotdbot.UpdateNewMessage) error {
		msg := m.Message
		senderID := msg.SenderID()
		if senderID != globalConfig.OwnerId && (database.IsBlacklisted(c.Me.Id, senderID) || database.IsBlacklisted(c.Me.Id, msg.ChatId)) {
			return gotdbot.EndGroups
		}

		go database.AddUserOrChat(c.Me.Id, msg.ChatId, msg.IsPrivate())
		return gotdbot.ContinueGroups
	}, nil, -1)
}

// isClientRunning checks whether a bot with the given ID is already registered and running.
func isClientRunning(manager *gotdbot.ClientManager, botID int64) bool {
	for _, client := range manager.GetClients() {
		if client.Me != nil && client.Me.Id == botID {
			return true
		}
	}
	return false
}
