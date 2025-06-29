package src

import (
	"fmt"
	"songBot/src/config"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

// startHandle responds to the /start command with a welcome message.
func startHandle(m *telegram.NewMessage) error {
	bot := m.Client.Me()
	name := m.Sender.FirstName

	go func() {
		if err := config.SaveUser(m.Sender.ID); err != nil {
			m.Client.Logger.Error("Save user error:", err)
		}
	}()

	response := fmt.Sprintf(`
👋 Hello <b>%s</b>!

🎧 <b>Welcome to %s</b> — your personal music downloader bot!

Supports: <b>Spotify</b>, <b>YouTube</b>, <b>Apple Music</b>, <b>SoundCloud</b>

<b>🔍 How to Use:</b>
• Send a song name or link directly  
• Inline: <code>@%s lofi mood</code>  
• Group: <code>/spotify &lt;url&gt;</code>

<b>🤖 Want Your Own Bot?</b>  
Clone it in 10 seconds using this guide:  
<a href="https://t.me/FallenProjects/131">Clone Your Bot via Token</a>

<b>🛑 Stop Your Clone:</b>  
Send <code>/stop</code> in <b>your cloned bot's private chat</b>

<b>🔗 Links:</b>  
🌟 <a href="https://t.me/FallenProjects">Support Channel</a>  
🛠️ <a href="https://github.com/AshokShau/SpTubeBot">Source Code</a>

Enjoy endless tunes! 🚀`, name, bot.FirstName, bot.Username)

	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.URL("💫 Fᴀʟʟᴇɴ Pʀᴏᴊᴇᴄᴛꜱ", "https://t.me/FallenProjects")).
		AddRow(telegram.Button.URL("📌 Clone Guide", "https://t.me/FallenProjects/131"))

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
