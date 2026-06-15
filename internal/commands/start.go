package commands

import (
	"fmt"

	"github.com/AshokShau/gotdbot"
)

func startHandler(c *gotdbot.Client, m *gotdbot.Message) error {
	username := c.Me.Usernames.EditableUsername
	text := `Welcome to <b>NoiNoi Bot</b> 🚀

Download songs & videos from Instagram, TikTok, YouTube, Spotify, and more — all in one place.

<b>How it works:</b>
• Send any link
• Get your download instantly

Fast. Simple. No hassle.

Join @FallenProjects for more bots & updates.`

	replyMarkup := &gotdbot.ReplyMarkupInlineKeyboard{
		Rows: [][]gotdbot.InlineKeyboardButton{
			{
				{
					Text:  "➕ Add me to your group",
					Type:  &gotdbot.InlineKeyboardButtonTypeUrl{Url: fmt.Sprintf("https://t.me/%s?startgroup=true", username)},
				},
			},
		},
	}

	_, err := m.ReplyText(c, text, &gotdbot.SendTextMessageOpts{
		ParseMode:   "HTML",
		ReplyMarkup: replyMarkup,
	})
	return err
}

func cloneHandler(c *gotdbot.Client, m *gotdbot.Message) error {
	return gotdbot.EndGroups
}
