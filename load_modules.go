package main

import (
	"regexp"
	"songBot/src"

	"github.com/amarnathcjd/gogram/telegram"
)

var urlPatterns = map[string]*regexp.Regexp{
	"spotify": regexp.MustCompile(`^(https?://)?(open\.spotify\.com/(track|playlist|album|artist)/[a-zA-Z0-9]+)(\?.*)?$`),
	//"soundcloud":    regexp.MustCompile(`^(https?://)?(www\.)?soundcloud\.com/[a-zA-Z0-9_-]+(/(sets)?/[a-zA-Z0-9_-]+)?(\?.*)?$`),
	"youtube":       regexp.MustCompile(`^(https?://)?(www\.)?(youtube\.com/watch\?v=|youtu\.be/)[a-zA-Z0-9_-]+(\?.*)?$`),
	"youtube_music": regexp.MustCompile(`^(https?://)?(music\.)?youtube\.com/(watch\?v=|playlist\?list=|)[a-zA-Z0-9_-]+(\?.*)?$`),
}

func filterURLChat(m *telegram.NewMessage) bool {
	text := m.Text()
	if m.IsCommand() || text == "" || m.IsForward() || m.Message.ViaBotID != 0 {
		return false
	}
	for _, pattern := range urlPatterns {
		if pattern.MatchString(text) {
			return true
		}
	}

	return m.IsPrivate()
}

func FilterOwner(m *telegram.NewMessage) bool {
	return m.SenderID() == 5938660179
}

func initFunc(c *telegram.Client) {
	_, _ = c.UpdatesGetState()

	c.On("command:start", src.StartHandle)
	c.On("command:ping", src.PingHandle)
	c.On("command:spotify", src.SpotifySearchSong)
	c.On("command:vid", src.YtVideoDL)
	c.On("command:privacy", src.PrivacyHandle)

	c.On("callback:spot_(.*)_(.*)", src.SpotifyHandlerCallback)
	c.AddRawHandler(&telegram.UpdateBotInlineSend{}, src.SpotifyInlineHandler)
	c.On(telegram.OnInline, src.SpotifyInlineSearch)

	c.On("command:ul", src.UploadHandle, telegram.FilterFunc(FilterOwner))
	c.On("command:dl", src.DownloadHandle, telegram.FilterFunc(FilterOwner))
	c.On("command:ver", src.GoGramVersion, telegram.FilterFunc(FilterOwner))

	c.On("message:*", src.SpotifySearchSong, telegram.FilterFunc(filterURLChat))
}
