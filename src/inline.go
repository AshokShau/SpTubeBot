package src

import (
	"fmt"
	"github.com/amarnathcjd/gogram/telegram"
	"os"
	"songBot/src/utils"
	"strings"
	"time"
)

// spotifyInlineSearch handles inline Spotify queries.
func spotifyInlineSearch(query *telegram.InlineQuery) error {
	q := strings.TrimSpace(query.Query)
	builder := query.Builder()

	if q == "" {
		builder.Article("❗️ No Query", "Please type something to search 🎵", "❗️ No query entered.")
		_, _ = query.Answer(builder.Results())
		return nil
	}

	searchData, err := utils.NewApiData(q).Search("15")
	if err != nil || len(searchData.Results) == 0 {
		builder.Article("⚠️ Error", "Failed to search Spotify.", "❌ Failed to search Spotify.")
		_, _ = query.Answer(builder.Results())
		return nil
	}

	for _, result := range searchData.Results {
		msg := fmt.Sprintf(
			`<b>🎧 Spotify Track</b>

<b>Name:</b> %s
<b>Artist:</b> %s
<b>Year:</b> %s

<b>Spotify ID:</b> <code>%s</code>`,
			result.Name, result.Artist, result.Year, result.ID,
		)
		builder.Article(
			fmt.Sprintf("%s - %s", result.Name, result.Artist),
			result.Year,
			msg,
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

// spotifyInlineHandler handles inline result selection.
func spotifyInlineHandler(update telegram.Update, client *telegram.Client) error {
	send := update.(*telegram.UpdateBotInlineSend)
	track, err := utils.NewApiData(send.ID).GetTrack()
	if err != nil {
		_, _ = client.EditMessage(&send.MsgID, 0, "❌ Spotify song not found.")
		return nil
	}

	dl, err := utils.NewDownload(*track)
	if err != nil {
		client.Logger.Warn("Invalid download:", err)
		_, _ = client.EditMessage(&send.MsgID, 0, "⚠️ Failed to download the song."+err.Error())
		return nil
	}

	audioFile, thumb, err := dl.Process()
	if err != nil || audioFile == "" {
		client.Logger.Warn("Process failed:", err)
		_, _ = client.EditMessage(&send.MsgID, 0, "⚠️ Failed to download the song.")
		return nil
	}

	file, err := os.Open(audioFile)
	if err != nil {
		client.Logger.Warn("[Inline] Failed to open audio file:", audioFile)
		_, _ = client.EditMessage(&send.MsgID, 0, "❌ Failed to open audio file.")
		return nil
	}
	defer file.Close()

	info, _ := file.Stat()
	if info.Size() == 0 {
		client.Logger.Warn("[Inline] File is empty:", audioFile)
		_, _ = client.EditMessage(&send.MsgID, 0, "❌ Audio file is empty.")
		return nil
	}

	progress := telegram.NewProgressManager(2).SetInlineMessage(client, &send.MsgID)
	caption := buildTrackCaption(track)
	options := prepareTrackMessageOptions(audioFile, thumb, track, progress)
	err = clientSendEditedMessage(client, &send.MsgID, caption, &options)
	if err != nil && strings.Contains(err.Error(), "MEDIA_EMPTY") {
		client.Logger.Warn("Retrying due to MEDIA_EMPTY...")
		time.Sleep(1 * time.Second)
		err = clientSendEditedMessage(client, &send.MsgID, caption, &options)
	}

	if err != nil {
		if wait := telegram.GetFloodWait(err); wait > 0 {
			client.Logger.Warn("Flood wait:", wait)
			time.Sleep(time.Duration(wait) * time.Second)
			err = clientSendEditedMessage(client, &send.MsgID, caption, &options)
		}
	}

	if err != nil {
		client.Logger.Warn("Edit failed:", err)
		_, _ = client.EditMessage(&send.MsgID, 0, "❌ Failed to send the song."+err.Error())
	}
	return err
}
