package src

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
	"songBot/src/utils"
)

// prepareTrackMessageOptions builds SendOptions for sending an audio track.
func prepareTrackMessageOptions(file any, thumb any, track *utils.TrackInfo, progress *telegram.ProgressManager) telegram.SendOptions {
	return telegram.SendOptions{
		ProgressManager: progress,
		Media:           file,
		Thumb:           thumb,
		Attributes:      buildAudioAttributes(track),
		Caption:         buildTrackCaption(track),
		MimeType:        "audio/mpeg",
		ReplyMarkup: telegram.NewKeyboard().AddRow(
			telegram.Button.URL("🎧 Fᴀʟʟᴇɴ Pʀᴏᴊᴇᴄᴛꜱ", "https://t.me/FallenProjects"),
		).Build(),
	}
}

// buildTrackCaption returns the caption string for a Spotify track.
func buildTrackCaption(track *utils.TrackInfo) string {
	return fmt.Sprintf("<b>🎵 %s - %d</b>\n<b>Artist:</b> %s", track.Name, track.Year, track.Artist)
}

// buildAudioAttributes returns audio metadata for sending audio files.
func buildAudioAttributes(track *utils.TrackInfo) []telegram.DocumentAttribute {
	return []telegram.DocumentAttribute{
		&telegram.DocumentAttributeAudio{
			Title:     track.Name,
			Performer: track.Artist + " @FallenProjects",
			Duration:  int32(track.Duration),
		},
	}
}

func clientSendEditedMessage(client *telegram.Client, msgID any, text string, opts *telegram.SendOptions) error {
	_, err := client.EditMessage(msgID, 0, text, opts)
	return err
}

// spotifySearchSong handles user input for searching Spotify tracks.
func spotifySearchSong(m *telegram.NewMessage) error {
	query := m.Text()
	if m.IsCommand() {
		query = m.Args()
	}
	if query == "" {
		_, err := m.Reply("❗ Please provide a song name or Spotify URL.")
		return err
	}

	api := utils.NewApiData(query)
	kb := telegram.NewKeyboard()

	if api.IsValid(query) {
		song, err := api.GetInfo()
		if err != nil || song == nil || len(song.Results) == 0 {
			_, _ = m.Reply("😢 Song not found.")
			return nil
		}

		for _, track := range song.Results {
			data := fmt.Sprintf("spot_%s_0", utils.EncodeURL(track.URL))
			kb.AddRow(telegram.Button.Data(fmt.Sprintf("%s - %s", track.Name, track.Artist), data))
		}
	} else {
		search, err := api.Search("5")
		if err != nil || len(search.Results) == 0 {
			_, _ = m.Reply("😔 No results found.")
			return nil
		}

		for _, track := range search.Results {
			data := fmt.Sprintf("spot_%s_%d", utils.EncodeURL(track.URL), m.SenderID())
			kb.AddRow(telegram.Button.Data(fmt.Sprintf("%s - %s", track.Name, track.Artist), data))
		}
	}

	_, err := m.Reply("<b>🎧 Select a song from below:</b>", telegram.SendOptions{
		ReplyMarkup: kb.Build(),
	})

	if err != nil {
		m.Client.Log.Error(err.Error())
		_, _ = m.Reply("⚠️ Too many results. Please use a direct track URL or reduce playlist size.")
	}

	return nil
}

// spotifyHandlerCallback handles callback queries from inline buttons.
func spotifyHandlerCallback(cb *telegram.CallbackQuery) error {
	data := cb.DataString()
	split1, split2 := strings.Index(data, "_"), strings.LastIndex(data, "_")
	if split1 == -1 || split2 == -1 || split1 == split2 {
		_, _ = cb.Answer("❌ Invalid selection.", &telegram.CallbackOptions{Alert: true})
		_, _ = cb.Delete()
		return nil
	}

	idEnc := data[split1+1 : split2]
	uid := data[split2+1:]
	if uid != "0" && uid != fmt.Sprint(cb.SenderID) {
		_, _ = cb.Answer("🚫 This action is not meant for you.", &telegram.CallbackOptions{Alert: true})
		return nil
	}

	_, _ = cb.Answer("🔄 Processing your request...", &telegram.CallbackOptions{Alert: true})

	url, err := utils.DecodeURL(idEnc)
	if err != nil {
		cb.Client.Logger.Warn("Failed to decode URL:", err.Error())
		_, _ = cb.Edit("❌ Failed to decode the URL.")
		return nil
	}

	track, err := utils.NewApiData(url).GetTrack()
	if err != nil {
		cb.Client.Logger.Warn("Failed to fetch track:", err.Error())
		_, _ = cb.Edit("❌ Could not fetch track details.")
		return nil
	}

	msg, _ := cb.Edit("⏬ Downloading the song...")
	audioFile, thumb, err := utils.NewDownload(*track).Process()
	if err != nil || audioFile == "" {
		cb.Client.Logger.Warn("Download/process failed:", err)
		_, _ = msg.Edit("⚠️ Failed to download the song.")
		return nil
	}

	msg, _ = msg.Edit("⏫ Uploading the song...")
	if matches := regexp.MustCompile(`https?://t\.me/([^/]+)/(\d+)`).FindStringSubmatch(audioFile); len(matches) == 3 {
		if id, err := strconv.Atoi(matches[2]); err == nil {
			if ref, err := msg.Client.GetMessageByID(matches[1], int32(id)); err == nil {
				audioFile, err = ref.Download(&telegram.DownloadOptions{FileName: ref.File.Name})
				if err != nil {
					_, _ = msg.Edit("⚠️ Failed to download file. " + err.Error())
					return nil
				}
			}
		}
	}
	defer func() {
		_ = os.Remove(audioFile)
	}()

	progress := telegram.NewProgressManager(4)
	progress.Edit(telegram.MediaDownloadProgress(msg, progress))

	time.Sleep(300 * time.Millisecond)
	opts := prepareTrackMessageOptions(audioFile, thumb, track, progress)
	if _, err = msg.Edit(buildTrackCaption(track), opts); err != nil {
		cb.Client.Logger.Warn("Send failed:", err.Error())
		_, _ = msg.Edit("❌ Failed to send the track.")
	}
	return nil
}

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

	audioFile, thumb, err := utils.NewDownload(*track).Process()
	if err != nil || audioFile == "" {
		client.Logger.Warn("Process failed:", err)
		_, _ = client.EditMessage(&send.MsgID, 0, "⚠️ Failed to download the song.")
		return nil
	}

	defer func() {
		_ = os.Remove(audioFile)
	}()

	progress := telegram.NewProgressManager(3).SetInlineMessage(client, &send.MsgID)
	caption := buildTrackCaption(track)
	options := prepareTrackMessageOptions(audioFile, thumb, track, progress)

	time.Sleep(300 * time.Millisecond)
	err = clientSendEditedMessage(client, &send.MsgID, caption, &options)
	if err != nil && strings.Contains(err.Error(), "MEDIA_EMPTY") {
		client.Logger.Warn("Retrying due to MEDIA_EMPTY...")
		time.Sleep(700 * time.Millisecond)
		err = clientSendEditedMessage(client, &send.MsgID, caption, &options)
	}

	if err != nil {
		client.Logger.Warn("Edit failed:", err)
		_, _ = client.EditMessage(&send.MsgID, 0, "❌ Failed to send the song."+err.Error())
	}
	return err
}
