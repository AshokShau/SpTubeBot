package src

import (
	"fmt"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

// StartHandle handles the /start command with a welcome message
func StartHandle(m *telegram.NewMessage) error {
	bot := m.Client.Me()

	response := fmt.Sprintf(
		`👋 Hey %s!
		
🎶 <b>Welcome to %s — your music download buddy!</b>

▶️ Just send a song name or drop a Spotify, YouTube, AppleMusic and SoundCloud link.
💬 Inline search: <code>@%s lofi mood</code>
📥 Group commands:
 ┗ /spotify url

Enjoy your music! 🔥`,
		m.Sender.FirstName,
		bot.FirstName,
		bot.Username,
	)

	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.URL("💫 Fᴀʟʟᴇɴ Pʀᴏᴊᴇᴄᴛs", "https://t.me/FallenProjects"))

	_, err := m.Reply(response, telegram.SendOptions{
		ReplyMarkup: keyboard.Build(),
	})

	return err
}

// PingHandle handles the /ping command with a latency check
func PingHandle(m *telegram.NewMessage) error {
	start := time.Now()

	// Send initial message and measure time
	sentMsg, err := m.Reply("Pinging...")
	if err != nil {
		return err
	}

	// Calculate and send response time
	latency := time.Since(start)
	_, err = sentMsg.Edit(fmt.Sprintf("<code>Pong!</code> <code>%s</code>", latency))

	return err
}
