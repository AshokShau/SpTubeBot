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
	"github.com/AshokShau/gotdbot/handlers"
)

var (
	startTime    = time.Now()
	manager      *gotdbot.ClientManager
	globalConfig *config.Config
)

func LoadCmd(d *gotdbot.Dispatcher, m *gotdbot.ClientManager, cfg *config.Config) {
	manager = m
	globalConfig = cfg
	d.AddHandler(handlers.NewCommand("ping", pingHandler))
	d.AddHandler(handlers.NewCommand("start", startHandler))
	d.AddHandler(handlers.NewCommand("help", startHandler))
	d.AddHandler(handlers.NewCommand("yt", ytCommandHandler))
	d.AddHandler(handlers.NewCommand("math", mathHandler))
	d.AddHandler(handlers.NewCommand("stop", stopHandler))
	d.AddHandler(handlers.NewCommand("stats", statsHandler))
	d.AddHandler(handlers.NewCommand("catbox", catboxHandler))
	d.AddHandler(handlers.NewCommand("tgm", catboxHandler))
	d.AddHandler(handlers.NewCommand("litterbox", litterboxHandler))
	d.AddHandler(handlers.NewCommand("block", blockHandler))
	d.AddHandler(handlers.NewCommand("unblock", unblockHandler))

	d.AddHandler(handlers.NewUpdateNewInlineQuery(nil, handleInlineQuery))
	d.AddHandler(handlers.NewUpdateNewGuestQuery(nil, handleGuestQuery))
	d.AddHandler(handlers.NewUpdateNewInlineCallbackQuery(nil, handleInlineCallbackQuery))

	d.AddHandler(handlers.NewUpdateNewMessage(func(u *gotdbot.UpdateNewMessage) bool {
		msg := u.Message
		if msg == nil {
			return false
		}

		if msg.IsCommand() {
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
	}, func(c *gotdbot.Client, ctx *gotdbot.Context) error {
		senderID := ctx.EffectiveMessage.SenderID()
		if senderID != globalConfig.OwnerId && (database.IsBlacklisted(c.Me.Id, senderID) || database.IsBlacklisted(c.Me.Id, ctx.EffectiveChatId)) {
			return gotdbot.EndGroups
		}

		text := ctx.EffectiveMessage.GetText()

		if httpx.YouTubeShortsPattern.MatchString(text) || httpx.YouTubePattern.MatchString(text) || httpx.YouTubePostPattern.MatchString(text) {
			return youtubeHandler(c, ctx)
		}

		for _, pattern := range httpx.SnapPatterns {
			if pattern.MatchString(text) {
				return snapHandler(c, ctx)
			}
		}

		for _, pattern := range httpx.MusicPatterns {
			if pattern.MatchString(text) {
				return musicHandler(c, ctx)
			}
		}

		return gotdbot.EndGroups
	}))

	d.AddHandler(handlers.NewUpdateNewCallbackQuery(func(u *gotdbot.UpdateNewCallbackQuery) bool {
		return u.DataString() == "clone_create"
	}, handleCloneCreate))

	d.AddHandler(handlers.NewUpdateNewCallbackQuery(func(u *gotdbot.UpdateNewCallbackQuery) bool {
		return u.DataString() == "clone_mybots"
	}, handleMyBots))

	d.AddHandler(handlers.NewUpdateNewCallbackQuery(func(u *gotdbot.UpdateNewCallbackQuery) bool {
		return strings.HasPrefix(u.DataString(), "bot_")
	}, handleBotManage))

	d.AddHandler(handlers.NewUpdateNewCallbackQuery(func(u *gotdbot.UpdateNewCallbackQuery) bool {
		return strings.HasPrefix(u.DataString(), "revoke_")
	}, handleBotRevoke))

	d.AddHandler(handlers.NewUpdateNewCallbackQuery(func(u *gotdbot.UpdateNewCallbackQuery) bool {
		return strings.HasPrefix(u.DataString(), "delete_")
	}, handleBotDelete))

	d.AddHandler(handlers.NewUpdateNewCallbackQuery(func(u *gotdbot.UpdateNewCallbackQuery) bool {
		return string(u.Payload.(*gotdbot.CallbackQueryPayloadData).Data) == "clone_back"
	}, handleCloneBack))

	d.AddHandler(handlers.NewUpdateManagedBot(nil, func(c *gotdbot.Client, ctx *gotdbot.Context) error {
		u := ctx.Update.UpdateManagedBot
		if u == nil {
			return nil
		}

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

		if isClientRunning(manager, u.BotUserId) {
			log.Info("Managed bot is already running, skipping start")
			return nil
		}

		clientConfig := gotdbot.DefaultClientConfig()
		clientConfig.Dispatcher = c.Dispatcher
		clientConfig.DatabaseDirectory = "db_" + strconv.FormatInt(u.BotUserId, 10)

		newBot, err := manager.RegisterClient(cfg.ApiId, cfg.ApiHash, botToken.Text, clientConfig)
		if err != nil {
			log.Error("Failed to start managed clone bot", "error", err)
			_, _ = c.SendTextMessage(u.UserId, "Your bot was created, but I failed to start the clone.", nil)
			return nil
		}

		username := newBot.Me.Usernames.EditableUsername
		_, _ = c.SendTextMessage(u.UserId, fmt.Sprintf("Your clone bot %s (@%s) is now up and running! 🎉", newBot.Me.FirstName, username), nil)
		return nil
	}))

	d.AddHandlerToGroup(handlers.NewUpdateNewMessage(nil, func(c *gotdbot.Client, ctx *gotdbot.Context) error {
		msg := ctx.EffectiveMessage
		senderID := msg.SenderID()
		if senderID != globalConfig.OwnerId && (database.IsBlacklisted(c.Me.Id, senderID) || database.IsBlacklisted(c.Me.Id, ctx.EffectiveChatId)) {
			return gotdbot.EndGroups
		}

		go database.AddUserOrChat(c.Me.Id, ctx.EffectiveChatId, msg.IsPrivate())
		return nil
	}), -1)

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
