package commands

import "github.com/AshokShau/gotdbot"

func startHandler(c *gotdbot.Client, ctx *gotdbot.Context) error {
	text := `Welcome to <b>NoiNoi Bot</b>! 🚀

I can help you download media from various platforms. Just send me a link, and I'll do the rest!

<b>Cloning Feature:</b>
You can create your own copy of this bot! Simply forward a message containing your bot token from @BotFather to this chat, and I'll start a clone for you.

<b>Stopping a Clone:</b>
If you want to stop your cloned bot and remove your token, just use the <code>/stop</code> command in your bot.

Join @FallenProjects for more cool bots and updates.`

	_, err := ctx.EffectiveMessage.ReplyText(c, text, &gotdbot.SendTextMessageOpts{
		ParseMode: "HTML",
	})
	if err != nil {
		return err
	}
	return gotdbot.EndGroups
}

func helpHandler(c *gotdbot.Client, ctx *gotdbot.Context) error {
	return gotdbot.EndGroups
}
