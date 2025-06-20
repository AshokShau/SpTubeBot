package src

import (
	"fmt"

	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

func StartHandle(m *telegram.NewMessage) error {
	me := m.Client.Me()
	text := fmt.Sprintf(
		"👋 Hey %s!\n\n"+
			"🎶 <b>Welcome to %s — your music download buddy!</b>\n\n"+
			"▶️ Just send a song name or drop a Spotify/YouTube link.\n"+
			"💬 Inline search: <code>@%s lofi mood</code>\n"+
			"📥 Group commands:\n"+
			" ┗ /spotify <url>\n"+
			" ┗ /vid <url>\n\n"+
			"Enjoy your music! 🔥",
		m.Sender.FirstName, me.FirstName, me.Username,
	)

	opts := telegram.SendOptions{
		ReplyMarkup: telegram.NewKeyboard().AddRow(
			telegram.Button.URL("💫 Fᴀʟʟᴇɴ Pʀᴏᴊᴇᴄᴛs", "https://t.me/FallenProjects"),
		).Build(),
	}
	_, _ = m.Reply(text, opts)
	return nil
}

func PingHandle(m *telegram.NewMessage) error {
	startTime := time.Now()
	sentMessage, _ := m.Reply("Pinging...")
	fmt.Println("Pong!")
	_, err := sentMessage.Edit(fmt.Sprintf("<code>Pong!</code> <code>%s</code>", time.Since(startTime).String()))
	return err
}
