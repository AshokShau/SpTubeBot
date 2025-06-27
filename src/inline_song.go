package src

import (
	"fmt"
	"strings"

	"songBot/src/utils"

	"github.com/amarnathcjd/gogram/telegram"
)

// SpotifyInlineSearch handles inline search queries for Spotify songs.
func SpotifyInlineSearch(query *telegram.InlineQuery) error {
	builder := query.Builder()
	queryText := strings.TrimSpace(query.Query)

	if queryText == "" {
		builder.Article("❗️ No Query", "Please type something to search for 🎵", "❗️ No query entered.")
		_, _ = query.Answer(builder.Results())
		return nil
	}

	searchData, err := utils.NewApiData(queryText).Search("20")
	if err != nil {
		builder.Article("⚠️ Error", err.Error(), "❌ Failed to search Spotify.")
		_, _ = query.Answer(builder.Results())
		return nil
	}

	if len(searchData.Results) == 0 {
		builder.Article("😢 No Results", "No results found for your query.", "🔍 Try another search.")
		_, _ = query.Answer(builder.Results())
		return nil
	}

	for _, result := range searchData.Results {
		title := fmt.Sprintf("%s - %s", result.Name, result.Artist)
		description := result.Year
		message := fmt.Sprintf(
			`<b>🎧 Spotify Track</b>

<b>Name:</b> %s
<b>Artist:</b> %s
<b>Year:</b> %s

<b>Spotify ID:</b> <code>%s</code>`,
			result.Name, result.Artist, result.Year, result.ID,
		)

		builder.Article(
			title,
			description,
			message,
			&telegram.ArticleOptions{
				ID: result.ID,
				ReplyMarkup: telegram.NewKeyboard().AddRow(
					telegram.Button.SwitchInline("🔁 Search Again", true, result.Artist),
				).Build(),
				Thumb: telegram.InputWebDocument{
					URL:      result.SmallCover,
					Size:     1500,
					MimeType: "image/jpeg",
				},
			},
		)
	}

	_, _ = query.Answer(builder.Results())
	return nil
}

// SpotifyInlineHandler handles the inline send event to fetch and send a Spotify song.
func SpotifyInlineHandler(update telegram.Update, client *telegram.Client) error {
	inlineSend := update.(*telegram.UpdateBotInlineSend)
	songID := inlineSend.ID
	client.Logger.Info("Received inline send event for song ID:", songID)
	track, err := utils.NewApiData("").GetTrack(songID)
	if err != nil {
		_, _ = client.EditMessage(&inlineSend.MsgID, 0, "❌ Spotify song not found.")
		return nil
	}

	downloader := utils.NewDownload(*track)
	audioFile, thumbnail, err := downloader.Process()
	if err != nil || audioFile == "" {
		if err != nil {
			client.Logger.Warn("Failed to download or process the song:", err.Error())
		}
		_, _ = client.EditMessage(&inlineSend.MsgID, 0, "⚠️ Failed to download or process the song.")
		return nil
	}

	caption := fmt.Sprintf("<b>🎵 %s - %d</b>\n<b>Artist:</b> %s", track.Name, track.Year, track.Artist)
	options := prepareTrackMessageOptions(
		track,
		audioFile,
		thumbnail,
		telegram.NewProgressManager(3).SetInlineMessage(client, &inlineSend.MsgID),
		caption,
	)

	_, err = client.EditMessage(&inlineSend.MsgID, 0, caption, &options)
	if err != nil {
		_, _ = client.EditMessage(&inlineSend.MsgID, 0, "❌ Failed to send the song.")
	}

	return nil
}
