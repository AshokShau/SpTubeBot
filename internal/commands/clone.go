package commands

import (
	"log"
	"noinoi/internal/database"
	"regexp"
	"strings"
	"time"

	"github.com/AshokShau/gotdbot"
)

var tokenRegex = regexp.MustCompile(`\d{8,11}:[A-Za-z0-9_-]{35}`)

func cloneHandler(c *gotdbot.Client, ctx *gotdbot.Context) error {
	msg := ctx.EffectiveMessage
	text := msg.GetText()
	log.Printf("Cloning %s", text)

	match := tokenRegex.FindString(text)
	if match == "" {
		_, _ = msg.ReplyText(c, "No valid bot token found in the forwarded message. Please forward a message that contains your bot token.", nil)
		return gotdbot.EndGroups
	}

	userId := ctx.EffectiveChatId
	botToken := match

	if manager == nil {
		_, _ = msg.ReplyText(c, "Internal error: ClientManager not initialized.", nil)
		return nil
	}
	go func() {
		reply, err := msg.ReplyText(c, "Cloning "+match, nil)
		clientConfig := gotdbot.DefaultClientConfig()
		clientConfig.Dispatcher = c.Dispatcher
		clientConfig.DatabaseDirectory = "db_" + strings.Split(botToken, ":")[0]

		newBot, err := manager.RegisterClient(globalConfig.ApiId, globalConfig.ApiHash, botToken, clientConfig)
		if err != nil {
			log.Printf("Failed to register new bot: %v", err)
			_, _ = msg.ReplyText(c, "Failed to register your bot. Is the token valid?", nil)
			return
		}

		err = database.SaveBot(database.BotInfo{
			UserId:    userId,
			BotToken:  botToken,
			CreatedAt: time.Now(),
		})
		if err != nil {
			log.Printf("Failed to save bot token to database: %v", err)
			_, _ = msg.ReplyText(c, "Bot started, but failed to save it to database.", nil)
		}

		me, _ := newBot.GetMe()
		_, err = reply.EditText(c, me.FirstName+" has been successfully started!", nil)
	}()

	return gotdbot.EndGroups
}

func stopHandler(c *gotdbot.Client, ctx *gotdbot.Context) error {
	userId := ctx.EffectiveChatId
	me, _ := c.GetMe()
	if me == nil {
		return nil
	}

	ownerId, ok := database.GetOwner(me.Id)
	if !ok || ownerId != userId {
		return nil
	}

	_, err := ctx.EffectiveMessage.ReplyText(c, "Stopping this bot and removing your token from database...", nil)
	if err != nil {
		return err
	}

	err = database.DeleteBot(userId)
	if err != nil {
		log.Printf("Failed to delete bot from database: %v", err)
	}

	go c.Close()
	return nil
}
