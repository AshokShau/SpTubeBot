package src

import (
	"fmt"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

// startHandle responds to the /start command with a welcome message.
func startHandle(m *telegram.NewMessage) error {
	bot := m.Client.Me()
	name := m.Sender.FirstName
	response := fmt.Sprintf(`
👋 Hello <b>%s</b>!

🎧 <b>Welcome to %s</b> — your personal music downloader bot!

Supports: <b>Spotify</b>, <b>YouTube</b>, <b>Apple Music</b>, <b>SoundCloud</b>

<b>🔍 How to Use:</b>
• Send a song name or link directly  
• Inline: <code>@%s lofi mood</code>  
• Group: <code>/spotify &lt;url&gt;</code>
• Playlist: <code>/playlist &lt;url&gt;</code>

<b>⚙️ Features:</b>
• Download songs from YouTube, Spotify, Apple Music, and SoundCloud  
• No ads  
• High quality audio  
• Seamless integration with Telegram groups

Enjoy endless tunes! 🚀`, name, bot.FirstName, bot.Username)

	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.URL("💫 Fᴀʟʟᴇɴ Pʀᴏᴊᴇᴄᴛꜱ", "https://t.me/FallenProjects")).
		AddRow(telegram.Button.URL("🛠️ Sᴏᴜʀᴄᴇ Cᴏᴅᴇ", "https://github.com/AshokShau/SpTubeBot"))
	_, err := m.Reply(response, telegram.SendOptions{
		ReplyMarkup: keyboard.Build(),
	})
	return err
}

// pingHandle responds to the /ping command with the bot's latency.
func pingHandle(m *telegram.NewMessage) error {
	start := time.Now()

	msg, err := m.Reply("⏱️ Pinging...")
	if err != nil {
		return err
	}

	latency := time.Since(start)
	_, err = msg.Edit(fmt.Sprintf("🏓 <b>Pong!</b> <code>%s</code>", latency))
	return err
}
