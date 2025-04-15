package main

import (
	"regexp"
	"songBot/src"

	"github.com/amarnathcjd/gogram/telegram"
)

func filterPmChat(m *telegram.NewMessage) bool {
	text := m.Text()
	if m.IsCommand() || text == "" || m.IsForward() || m.Message.ViaBotID != 0 {
		return false
	}

	urlRegex := regexp.MustCompile(`(?i)\b(?:https?://|www\.)\S+\b`)
	isUrl := urlRegex.MatchString(text)

	if len(text) > 50 && !isUrl {
		return false
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

	c.On("callback:spot_(.*)_(.*)", src.SpotifyHandlerCallback)
	c.AddRawHandler(&telegram.UpdateBotInlineSend{}, src.SpotifyInlineHandler)
	c.On(telegram.OnInline, src.SpotifyInlineSearch)

	c.On("command:ul", src.UploadHandle, telegram.FilterFunc(FilterOwner))
	c.On("command:dl", src.DownloadHandle, telegram.FilterFunc(FilterOwner))
	c.On("command:ver", src.GoGramVersion, telegram.FilterFunc(FilterOwner))

	c.On("message:*", src.SpotifySearchSong, telegram.FilterFunc(filterPmChat))
}
