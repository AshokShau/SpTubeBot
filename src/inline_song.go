package src

import (
	"fmt"
	"strings"

	"songBot/src/utils"

	"github.com/amarnathcjd/gogram/telegram"
)

// SpotifyInlineSearch handles inline search queries.
func SpotifyInlineSearch(query *telegram.InlineQuery) error {
	builder := query.Builder()
	args := query.Query

	if args == "" {
		builder.Article("No query", "Please enter a query to search for.", "No query")
		_, _ = query.Answer(builder.Results())
		return nil
	}

	searchData, err := utils.NewApiData(args).Search("20")
	if err != nil {
		builder.Article("Error", err.Error(), "Error")
		_, _ = query.Answer(builder.Results())
		return nil
	}

	if len(searchData.Results) == 0 {
		builder.Article("No results", "No results found.", "No results")
		_, _ = query.Answer(builder.Results())
		return nil
	}

	for _, result := range searchData.Results {
		builder.Article(
			fmt.Sprintf("%s - %s", result.Name, result.Artist),
			result.Year,
			fmt.Sprintf("<b>Spotify Song - Searching...</b>\n\n<b>Name:</b> %s\n<b>Artist:</b> %s\n<b>Year:</b> %s\n\n<b>Spotify ID:</b> <code>%s</code>", result.Name, result.Artist, result.Year, result.ID),
			&telegram.ArticleOptions{
				ID: result.ID,
				ReplyMarkup: telegram.NewKeyboard().AddRow(
					telegram.Button.SwitchInline("Search Again", true, result.Artist),
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

// SpotifyInlineHandler sends the song to the user.
func SpotifyInlineHandler(u telegram.Update, c *telegram.Client) error {
	i := u.(*telegram.UpdateBotInlineSend)
	if strings.Contains(i.Query, "pin") {
		return nil
	}

	songID := i.ID

	// Fetch track details from Spotify
	track, err := utils.NewApiData("").GetTrack(songID)
	if err != nil {
		_, _ = c.EditMessage(&i.MsgID, 0, "Spotify song not found.")
		return nil
	}

	downloader := utils.NewDownload(*track)
	audioFile, thumbnail, err := downloader.Process()
	if err != nil {
		_, _ = c.EditMessage(&i.MsgID, 0, "Failed to download the song.")
		return nil
	}

	if audioFile == "" {
		_, _ = c.EditMessage(&i.MsgID, 0, "Failed to process the song.")
		return nil
	}

	caption := fmt.Sprintf("<b> %s- %d</b>\n<b>Artist:</b> %s", track.Name, track.Year, track.Artist)
	options := prepareTrackMessageOptions(track, audioFile, thumbnail, telegram.NewProgressManager(3).SetInlineMessage(c, &i.MsgID), caption)
	_, err = c.EditMessage(&i.MsgID, 0, caption, &options)
	if err != nil {
		_, _ = c.EditMessage(&i.MsgID, 0, "Failed to send the song.")
		return nil
	}

	return nil
}
