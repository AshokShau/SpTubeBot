package src

import (
	"songBot/src/utils"

	"github.com/amarnathcjd/gogram/telegram"
)

// filterURLChat handles messages that are not commands but contain supported URLs or are private
func filterURLChat(m *telegram.NewMessage) bool {
	text := m.Text()

	if m.IsCommand() || text == "" || m.IsForward() || m.Message.ViaBotID == m.Client.Me().ID {
		return false
	}

	for _, pattern := range utils.UrlPatterns {
		if pattern.MatchString(text) {
			return true
		}
	}

	return m.IsPrivate()
}

// FilterOwner allows only bot owner access to sensitive commands
func FilterOwner(m *telegram.NewMessage) bool {
	return m.SenderID() == 5938660179
}

// InitFunc initializes the bot and registers all command, message, and callback handlers
func InitFunc(c *telegram.Client) {
	_, _ = c.UpdatesGetState()
	// Public commands
	c.On("command:start", startHandle)
	c.On("command:ping", pingHandle)
	c.On("command:spotify", spotifySearchSong)
	c.On("command:privacy", privacyHandle)
	c.On("command:playlist", zipHandle)

	// Inline query and inline result handler
	c.On(telegram.OnInline, spotifyInlineSearch)
	c.AddRawHandler(&telegram.UpdateBotInlineSend{}, spotifyInlineHandler)

	// Spotify inline button callback
	c.On("callback:spot_(.*)_(.*)", spotifyHandlerCallback)

	// Owner-only commands
	c.On("command:ul", uploadHandle, telegram.FilterFunc(FilterOwner))
	c.On("command:dl", downloadHandle, telegram.FilterFunc(FilterOwner))

	// Fallback message handler for plain URLs or private messages
	c.On("message:*", spotifySearchSong, telegram.FilterFunc(filterURLChat))

}
