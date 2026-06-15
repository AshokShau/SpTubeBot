package commands

import (
	"fmt"
	"net/url"
	"noinoi/internal/database"

	"github.com/AshokShau/gotdbot"
)

func handleCloneCreate(c *gotdbot.Client, cb *gotdbot.UpdateNewCallbackQuery) error {
	userId := cb.SenderUserId

	bots, err := database.GetBotsByOwner(userId)
	if err != nil {
		c.Logger.Warn("Failed to fetch bots for user", "user_id", userId, "error", err)
		_ = cb.Answer(c, 0, true, "Failed to fetch your bots.", "")
		return gotdbot.EndGroups
	}

	if len(bots) >= 5 {
		_ = cb.Answer(c, 0, true, "You have reached the maximum limit of 5 bots. Please delete an existing bot before creating a new one.", "")
		return gotdbot.EndGroups
	}

	_ = cb.Answer(c, 0, false, "loading..", "")

	botLink := fmt.Sprintf("https://t.me/newbot/%s/%s?name=%s", c.Me.Usernames.EditableUsername, "noinoi_clone", url.QueryEscape("Downloader Bot Clone"))
	text := fmt.Sprintf("Almost there! Click the button below to confirm and create your bot via Telegram.\n\nOnce you create it, I will automatically start it for you!")

	replyMarkup := &gotdbot.ReplyMarkupInlineKeyboard{
		Rows: [][]gotdbot.InlineKeyboardButton{
			{
				{Text: "🚀 Create Bot", Type: &gotdbot.InlineKeyboardButtonTypeUrl{Url: botLink}},
			},
		},
	}

	_, err = cb.EditMessageText(c, text, &gotdbot.EditTextMessageOpts{ReplyMarkup: replyMarkup})
	if err != nil {
		return err
	}

	return gotdbot.EndGroups
}

func stopHandler(c *gotdbot.Client, m *gotdbot.Message) error {
	userId := m.ChatId
	ownerId, ok := database.GetOwner(c.Me.Id)
	if !ok || ownerId != userId {
		return nil
	}

	_, err := m.ReplyText(c, "Stopping this bot and removing your token from database...", nil)
	if err != nil {
		return err
	}

	err = database.DeleteBot(c.Me.Id)
	if err != nil {
		c.Logger.Warn("Failed to delete bot", "bot_id", c.Me.Id, "error", err)
	}

	go c.Close()
	return nil
}
