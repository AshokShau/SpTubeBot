package commands

import (
	"noinoi/internal/config"
	"noinoi/internal/httpx"
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
	d.AddHandler(handlers.NewCommand("help", helpHandler))
	d.AddHandler(handlers.NewCommand("yt", ytCommandHandler))
	d.AddHandler(handlers.NewCommand("math", mathHandler))
	d.AddHandler(handlers.NewCommand("stop", stopHandler))

	d.AddHandler(handlers.NewUpdateNewMessage(func(u *gotdbot.UpdateNewMessage) bool {
		msg := u.Message
		if msg == nil || msg.ForwardInfo == nil {
			return false
		}

		if !msg.IsPrivate() {
			return false
		}

		origin := msg.ForwardInfo.Origin
		if origin == nil {
			return false
		}

		var senderId int64
		switch o := origin.(type) {
		case *gotdbot.MessageOriginUser:
			senderId = o.SenderUserId
		}

		return senderId == 93372553
	}, cloneHandler))

	d.AddHandler(handlers.NewUpdateNewInlineQuery(nil, handleInlineQuery))
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

		if httpx.YouTubeShortsPattern.MatchString(text) || httpx.YouTubePattern.MatchString(text) {
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
		text := ctx.EffectiveMessage.GetText()

		if httpx.YouTubeShortsPattern.MatchString(text) || httpx.YouTubePattern.MatchString(text) {
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
}
