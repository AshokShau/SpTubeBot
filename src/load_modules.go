package src

import (
	"regexp"

	"github.com/amarnathcjd/gogram/telegram"
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

func filterClone(m *telegram.NewMessage) bool {
	if !m.IsForward() {
		return false
	}

	fwd := m.Message.FwdFrom
	if fwd == nil || fwd.FromID == nil {
		return false
	}

	switch peer := fwd.FromID.(type) {
	case *telegram.PeerUser:
		if peer.UserID != 93372553 {
			return false
		}
	default:
		return false
	}

	text := m.Text()
	if m.IsCommand() || text == "" {
		return false
	}

	var tokenRegex = regexp.MustCompile(`\b\d{6,}:[\w-]{30,}\b`)
	match := tokenRegex.FindString(text)

	return match != ""
}

// InitFunc initializes the bot and registers all command, message, and callback handlers
func InitFunc(c *telegram.Client) {
	_, _ = c.UpdatesGetState()
	// Public commands
	c.On("command:start", startHandle)
	c.On("command:ping", pingHandle)
	c.On("command:spotify", spotifySearchSong)
	c.On("command:privacy", privacyHandle)
	c.On("command:stop", stopHandler)

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

	// Clone
	c.On("message:*", cloneHandle, telegram.FilterFunc(filterClone))
}
