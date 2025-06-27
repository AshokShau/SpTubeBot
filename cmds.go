package main

import (
	"regexp"

	"github.com/amarnathcjd/gogram/telegram"
	"songBot/src"
)

// Supported music platform URL patterns
var urlPatterns = map[string]*regexp.Regexp{
	"spotify":       regexp.MustCompile(`^(https?://)?(open\.spotify\.com/(track|playlist|album|artist)/[a-zA-Z0-9]+)(\?.*)?$`),
	"youtube":       regexp.MustCompile(`^(https?://)?(www\.)?(youtube\.com/watch\?v=|youtu\.be/)[\w-]+(\?.*)?$`),
	"youtube_music": regexp.MustCompile(`^(https?://)?(music\.)?youtube\.com/(watch\?v=|playlist\?list=)[\w-]+(\?.*)?$`),
	"soundcloud":    regexp.MustCompile(`^(https?://)?(www\.)?soundcloud\.com/[\w-]+(/[\w-]+)?(/sets/[\w-]+)?(\?.*)?$`),
	"apple_music":   regexp.MustCompile(`^(https?://)?(music|geo)\.apple\.com/[a-z]{2}/(album|playlist|song)/[^/]+/\d+(\?i=\d+)?(\?.*)?$`),
}

// filterURLChat handles messages that are not commands but contain supported URLs or are private
func filterURLChat(m *telegram.NewMessage) bool {
	text := m.Text()

	if m.IsCommand() || text == "" || m.IsForward() || m.Message.ViaBotID == m.Client.Me().ID {
		return false
	}

	for _, pattern := range urlPatterns {
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

// initFunc initializes the bot and registers all command, message, and callback handlers
func initFunc(c *telegram.Client) {
	_, _ = c.UpdatesGetState()

	// Public commands
	c.On("command:start", src.StartHandle)
	c.On("command:ping", src.PingHandle)
	c.On("command:spotify", src.SpotifySearchSong)
	c.On("command:privacy", src.PrivacyHandle)

	// Inline query and inline result handler
	c.On(telegram.OnInline, src.SpotifyInlineSearch)
	c.AddRawHandler(&telegram.UpdateBotInlineSend{}, src.SpotifyInlineHandler)

	// Spotify inline button callback
	c.On("callback:spot_(.*)_(.*)", src.SpotifyHandlerCallback)

	// Owner-only commands
	c.On("command:ul", src.UploadHandle, telegram.FilterFunc(FilterOwner))
	c.On("command:dl", src.DownloadHandle, telegram.FilterFunc(FilterOwner))
	c.On("command:ver", src.GoGramVersion, telegram.FilterFunc(FilterOwner))

	// Fallback message handler for plain URLs or private messages
	c.On("message:*", src.SpotifySearchSong, telegram.FilterFunc(filterURLChat))
}
