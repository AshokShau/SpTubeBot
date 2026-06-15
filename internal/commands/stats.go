package commands

import (
	"fmt"
	"noinoi/internal/database"
	"runtime"
	"time"

	"github.com/AshokShau/gotdbot"
)

func statsHandler(c *gotdbot.Client, m *gotdbot.Message) error {
	botId := c.Me.Id
	ownerId, ok := database.GetOwner(botId)

	senderId := m.SenderID()
	isMainOwner := senderId == globalConfig.OwnerId
	isCloneOwner := ok && senderId == ownerId

	mainBotId := database.ParseBotId(globalConfig.Token)
	isMainBot := botId == mainBotId

	if isMainBot {
		if !isMainOwner {
			return nil
		}
	} else {
		if !isCloneOwner && !isMainOwner {
			return nil
		}
	}

	var statsText string

	if isMainBot {
		totalBots, totalUsers, totalChats, totalSuccess, totalFailed, err := database.GetGlobalStats()
		if err != nil {
			_, _ = m.ReplyText(c, fmt.Sprintf("Error fetching global stats: %v", err), nil)
			return nil
		}

		statsText = fmt.Sprintf(
			"📊 <b>Global Bot Statistics</b>\n\n"+
				"🤖 <b>Total Bots:</b> <code>%d</code>\n"+
				"👤 <b>Total Users:</b> <code>%d</code>\n"+
				"💬 <b>Total Chats:</b> <code>%d</code>\n"+
				"✅ <b>Downloads (Success):</b> <code>%d</code>\n"+
				"❌ <b>Downloads (Failed):</b> <code>%d</code>\n\n",
			totalBots, totalUsers, totalChats, totalSuccess, totalFailed,
		)
	} else {
		stats, err := database.GetBotStats(botId)
		if err != nil {
			_, _ = m.ReplyText(c, fmt.Sprintf("Error fetching bot stats: %v", err), nil)
			return nil
		}

		statsText = fmt.Sprintf(
			"📊 <b>Bot Statistics</b>\n\n"+
				"👤 <b>Total Users:</b> <code>%d</code>\n"+
				"💬 <b>Total Chats:</b> <code>%d</code>\n"+
				"✅ <b>Downloads (Success):</b> <code>%d</code>\n"+
				"❌ <b>Downloads (Failed):</b> <code>%d</code>\n\n",
			len(stats.Users), len(stats.Chats), stats.SuccessDownloads, stats.FailedDownloads,
		)
	}

	var mStats runtime.MemStats
	runtime.ReadMemStats(&mStats)

	uptime := time.Since(startTime).Truncate(time.Second)

	statsText += fmt.Sprintf(
		"⚙️ <b>System Information</b>\n\n"+
			"🚀 <b>Uptime:</b> <code>%s</code>\n"+
			"📦 <b>Go Version:</b> <code>%s</code>\n"+
			"🧵 <b>Goroutines:</b> <code>%d</code>\n"+
			"💾 <b>RAM (Alloc):</b> <code>%.2f MB</code>\n"+
			"💾 <b>RAM (Sys):</b> <code>%.2f MB</code>\n",
		uptime, runtime.Version(), runtime.NumGoroutine(),
		float64(mStats.Alloc)/1024/1024,
		float64(mStats.Sys)/1024/1024,
	)

	_, err := m.ReplyText(c, statsText, &gotdbot.SendTextMessageOpts{ParseMode: "HTML"})
	return err
}
