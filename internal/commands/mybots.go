package commands

import (
	"fmt"
	"noinoi/internal/database"
	"strconv"
	"strings"
	"time"

	"github.com/AshokShau/gotdbot"
)

func handleMyBots(c *gotdbot.Client, cb *gotdbot.UpdateNewCallbackQuery) error {
	userId := cb.SenderUserId
	if userId != globalConfig.OwnerId && database.IsBlacklisted(c.Me.Id, userId) {
		return gotdbot.EndGroups
	}

	bots, err := database.GetBotsByOwner(userId)
	if err != nil {
		_ = cb.Answer(c, 0, true, "Failed to fetch your bots.", "")
		return err
	}

	if len(bots) == 0 {
		_ = cb.Answer(c, 0, true, "You don't have any cloned bots yet.", "")
		return nil
	}

	_ = cb.Answer(c, 0, true, "loading...", "")

	var rows [][]gotdbot.InlineKeyboardButton
	for _, b := range bots {
		botUser, err := c.GetUser(b.BotId)
		var btnText string

		if err == nil && botUser != nil && botUser.Usernames != nil && len(botUser.Usernames.ActiveUsernames) > 0 {
			btnText = "@" + botUser.Usernames.ActiveUsernames[0]
		} else if err == nil && botUser != nil {
			btnText = botUser.FirstName
		} else {
			btnText = fmt.Sprintf("Bot ID: %d", b.BotId)
		}

		rows = append(rows, []gotdbot.InlineKeyboardButton{
			{Text: btnText, Type: &gotdbot.InlineKeyboardButtonTypeCallback{Data: []byte("bot_" + strconv.FormatInt(b.BotId, 10))}},
		})
	}

	rows = append(rows, []gotdbot.InlineKeyboardButton{
		{Text: "🔙 Back", Type: &gotdbot.InlineKeyboardButtonTypeCallback{Data: []byte("clone_back")}},
	})

	replyMarkup := &gotdbot.ReplyMarkupInlineKeyboard{Rows: rows}

	_, err = cb.EditMessageText(c, "Here are your cloned bots. Select one to manage:", &gotdbot.EditTextMessageOpts{ReplyMarkup: replyMarkup})
	if err != nil {
		return err
	}

	return nil
}

func handleBotManage(c *gotdbot.Client, cb *gotdbot.UpdateNewCallbackQuery) error {
	senderID := cb.SenderUserId
	if senderID != globalConfig.OwnerId && database.IsBlacklisted(c.Me.Id, senderID) {
		return gotdbot.EndGroups
	}

	data := cb.DataString()
	botIdStr := strings.TrimPrefix(data, "bot_")
	botId, err := strconv.ParseInt(botIdStr, 10, 64)
	if err != nil {
		return err
	}

	ownerId, ok := database.GetOwner(botId)
	if !ok || ownerId != cb.SenderUserId {
		_ = cb.Answer(c, 0, true, "You are not the owner of this bot.", "")
		return nil
	}

	_ = cb.Answer(c, 0, true, "loading...", "")
	botUser, err := c.GetUser(botId)
	var botName string
	if err == nil && botUser != nil {
		botName = botUser.FirstName
	} else {
		botName = fmt.Sprintf("ID: %d", botId)
	}

	replyMarkup := &gotdbot.ReplyMarkupInlineKeyboard{
		Rows: [][]gotdbot.InlineKeyboardButton{
			{
				{Text: "🔄 Revoke Token", Type: &gotdbot.InlineKeyboardButtonTypeCallback{Data: []byte("revoke_" + botIdStr)}},
				{Text: "⏹ Stop & Delete", Type: &gotdbot.InlineKeyboardButtonTypeCallback{Data: []byte("delete_" + botIdStr)}},
			},
			{
				{Text: "🔙 Back to My Bots", Type: &gotdbot.InlineKeyboardButtonTypeCallback{Data: []byte("clone_mybots")}},
			},
		},
	}

	text := fmt.Sprintf("Managing <b>%s</b>:\n\n• <b>Revoke Token</b>: Generates a new token and restarts the bot.\n• <b>Stop & Delete (from DB)</b>: Stops the bot permanently.", botName)
	_, err = cb.EditMessageText(c, text, &gotdbot.EditTextMessageOpts{
		ParseMode:   "HTML",
		ReplyMarkup: replyMarkup,
	})
	return err
}

func handleBotRevoke(c *gotdbot.Client, cb *gotdbot.UpdateNewCallbackQuery) error {
	if cb.SenderUserId != globalConfig.OwnerId && database.IsBlacklisted(c.Me.Id, cb.SenderUserId) {
		return gotdbot.EndGroups
	}

	data := cb.DataString()
	botIdStr := strings.TrimPrefix(data, "revoke_")
	botId, err := strconv.ParseInt(botIdStr, 10, 64)
	if err != nil {
		return err
	}

	ownerId, ok := database.GetOwner(botId)
	if !ok || ownerId != cb.SenderUserId {
		_ = cb.Answer(c, 0, true, "You are not the owner of this bot.", "")
		return nil
	}

	_ = cb.Answer(c, 0, true, "revoking token ...", "")

	newToken, err := c.GetManagedBotToken(botId, &gotdbot.GetManagedBotTokenOpts{Revoke: true})
	if err != nil {
		_, _ = cb.EditMessageText(c, fmt.Sprintf("Failed to revoke token: %v", err), nil)
		return nil
	}

	for _, cloneClient := range manager.GetClients() {
		if cloneClient.Me != nil && cloneClient.Me.Id == botId {
			cloneClient.Close()
			break
		}
	}

	err = database.SaveBot(database.BotInfo{
		BotId:     botId,
		OwnerId:   cb.SenderUserId,
		BotToken:  newToken.Text,
		CreatedAt: time.Now(),
	})

	clientConfig := gotdbot.DefaultClientConfig()
	clientConfig.DatabaseDirectory = "db_" + botIdStr

	_, err = manager.RegisterClient(globalConfig.ApiId, globalConfig.ApiHash, newToken.Text, clientConfig)
	if err != nil {
		_, _ = cb.EditMessageText(c, fmt.Sprintf("Token revoked but failed to start new client: %v", err), nil)
		return nil
	}

	_, _ = cb.EditMessageText(c, "Token revoked and bot restarted successfully.", nil)
	return nil
}

func handleBotDelete(c *gotdbot.Client, cb *gotdbot.UpdateNewCallbackQuery) error {
	if cb.SenderUserId != globalConfig.OwnerId && database.IsBlacklisted(c.Me.Id, cb.SenderUserId) {
		return gotdbot.EndGroups
	}

	data := cb.DataString()
	botIdStr := strings.TrimPrefix(data, "delete_")
	botId, err := strconv.ParseInt(botIdStr, 10, 64)
	if err != nil {
		return err
	}

	ownerId, ok := database.GetOwner(botId)
	if !ok || ownerId != cb.SenderUserId {
		_ = cb.Answer(c, 0, true, "You are not the owner of this bot.", "")
		return nil
	}

	_ = cb.Answer(c, 0, true, "deleting token & Closes the TDLib instance.", "")
	_, _ = cb.EditMessageText(c, fmt.Sprintf("Closed the TDLib instance: %s", botIdStr), nil)

	for _, cloneClient := range manager.GetClients() {
		if cloneClient.Me != nil && cloneClient.Me.Id == botId {
			cloneClient.Close()
			break
		}
	}

	_ = database.DeleteBot(botId)
	return nil
}

func handleCloneBack(c *gotdbot.Client, cb *gotdbot.UpdateNewCallbackQuery) error {
	if cb.SenderUserId != globalConfig.OwnerId && database.IsBlacklisted(c.Me.Id, cb.SenderUserId) {
		return gotdbot.EndGroups
	}

	_ = cb.Answer(c, 0, false, "", "")

	text := `Welcome to <b>NoiNoi Bot</b>! 🚀

I can help you download media from various platforms. Just send me a link, and I'll do the rest!

<b>Cloning Feature:</b>
You can create your own copy of this bot! Click the "Create Bot" button below.

Join @FallenProjects for more cool bots and updates.`

	replyMarkup := &gotdbot.ReplyMarkupInlineKeyboard{
		Rows: [][]gotdbot.InlineKeyboardButton{
			{
				{Text: "➕ Create Bot", Type: &gotdbot.InlineKeyboardButtonTypeCallback{Data: []byte("clone_create")}},
				{Text: "🤖 My Bots", Type: &gotdbot.InlineKeyboardButtonTypeCallback{Data: []byte("clone_mybots")}},
			},
		},
	}

	_, _ = cb.EditMessageText(c, text, &gotdbot.EditTextMessageOpts{ParseMode: "HTML", ReplyMarkup: replyMarkup})
	return nil
}
