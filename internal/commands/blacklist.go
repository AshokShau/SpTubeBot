package commands

import (
	"errors"
	"fmt"
	"noinoi/internal/database"
	"strconv"
	"strings"

	"github.com/AshokShau/gotdbot"
)

// getTargetUserID resolves a target user ID from a reply or command arguments.
func getTargetUserID(c *gotdbot.Client, m *gotdbot.Message) (int64, error) {
	if m.ReplyToMessageID() != 0 {
		return resolveFromReply(c, m)
	}

	args := strings.Fields(Args(m))
	if len(args) > 0 {
		return resolveFromArg(c, args[0])
	}

	return 0, errors.New("no target specified: reply to a message or provide a user ID/username")
}

// resolveFromReply extracts the sender ID from the replied-to message.
func resolveFromReply(c *gotdbot.Client, m *gotdbot.Message) (int64, error) {
	replyMsg, err := m.GetRepliedMessage(c)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch replied message: %w", err)
	}

	userID := replyMsg.SenderID()
	if userID == 0 {
		return 0, errors.New("replied message has no identifiable sender")
	}

	return userID, nil
}

// resolveFromArg parses a user ID or @username from a raw argument string.
func resolveFromArg(c *gotdbot.Client, arg string) (int64, error) {
	if id, err := strconv.ParseInt(arg, 10, 64); err == nil {
		return id, nil
	}

	return resolveUsername(c, arg)
}

// resolveUsername looks up a Telegram username and returns its chat ID.
func resolveUsername(c *gotdbot.Client, username string) (int64, error) {
	username = strings.TrimPrefix(username, "@")
	if username == "" {
		return 0, errors.New("username cannot be empty")
	}

	chat, err := c.SearchPublicChat(username)
	if err != nil {
		return 0, fmt.Errorf("username lookup failed for %q: %w", username, err)
	}
	if chat == nil {
		return 0, fmt.Errorf("no user found for username %q", username)
	}

	return chat.Id, nil
}

func Args(m *gotdbot.Message) string {
	text := m.Text()
	fields := strings.Fields(text)
	if len(fields) < 2 {
		return ""
	}

	firstFieldEnd := strings.Index(text, fields[0]) + len(fields[0])
	return strings.TrimSpace(text[firstFieldEnd:])
}

func canManageBlacklist(botID, senderID int64) bool {
	if senderID == globalConfig.OwnerId {
		return true
	}
	ownerID, ok := database.GetOwner(botID)
	return ok && senderID == ownerID
}

func blockHandler(c *gotdbot.Client, m *gotdbot.Message) error {
	botID := c.Me.Id
	senderID := m.SenderID()
	if !canManageBlacklist(botID, senderID) {
		return nil
	}

	targetID, err := getTargetUserID(c, m)
	if err != nil {
		_, err = m.ReplyText(c, fmt.Sprintf("Error: %v", err), nil)
		return err
	}

	if targetID == senderID {
		_, err = m.ReplyText(c, "You cannot block yourself.", nil)
		return err
	}

	if targetID == globalConfig.OwnerId {
		_, err = m.ReplyText(c, "You cannot block the bot owner.", nil)
		return err
	}

	err = database.BlacklistEntity(botID, targetID)
	if err != nil {
		_, err = m.ReplyText(c, fmt.Sprintf("Failed to block: %v", err), nil)
		return err
	}

	// Send one-time message
	var blockMsg string
	if targetID < 0 {
		blockMsg = "🚫 This chat has been restricted from using this bot.\n\nIf you think this is a mistake or need help, contact support:\nhttps://t.me/FallenSupport"
	} else {
		blockMsg = "🚫 You have been blocked from using this bot.\n\nIf you believe this is a mistake or have any questions, feel free to contact support:\nhttps://t.me/FallenSupport"
	}

	_, _ = c.SendTextMessage(targetID, blockMsg, nil)
	_, err = m.ReplyText(c, fmt.Sprintf("Successfully blocked ID: <code>%d</code> on this bot.", targetID), &gotdbot.SendTextMessageOpts{ParseMode: "HTML"})
	return err
}

func unblockHandler(c *gotdbot.Client, m *gotdbot.Message) error {
	botID := c.Me.Id
	senderID := m.SenderID()
	if !canManageBlacklist(botID, senderID) {
		return nil
	}

	targetID, err := getTargetUserID(c, m)
	if err != nil {
		_, err = m.ReplyText(c, fmt.Sprintf("Error: %v", err), nil)
		return err
	}

	err = database.WhitelistEntity(botID, targetID)
	if err != nil {
		_, err = m.ReplyText(c, fmt.Sprintf("Failed to unblock: %v", err), nil)
		return err
	}

	_, err = m.ReplyText(c, fmt.Sprintf("Successfully unblocked ID: <code>%d</code> on this bot.", targetID), &gotdbot.SendTextMessageOpts{ParseMode: "HTML"})
	return err
}
