package src

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
	"songBot/src/utils"
)

// SpotifySearchSong handles song search requests from user messages.
func SpotifySearchSong(m *telegram.NewMessage) error {
	songName := m.Text()
	if m.IsCommand() {
		songName = m.Args()
	}

	if songName == "" {
		_, err := m.Reply("❗ Please provide a song name or Spotify URL.")
		return err
	}

	sp := utils.NewApiData(songName)
	kb := telegram.NewKeyboard()

	if sp.IsValid(songName) {
		song, err := sp.GetInfo(songName)
		if err != nil {
			_, _ = m.Reply(fmt.Sprintf("⚠️ Error: %v", err))
			return nil
		}

		if song == nil || len(song.Results) == 0 {
			_, _ = m.Reply("😢 Song not found.")
			return nil
		}

		for _, track := range song.Results {
			data := fmt.Sprintf("spot_%s_0", utils.EncodeURL(track.URL))
			kb.AddRow(telegram.Button.Data(fmt.Sprintf("%s - %s", track.Name, track.Artist), data))
		}

		_, err = m.Reply("<b>🎧 Select a song from below:</b>", telegram.SendOptions{
			ReplyMarkup: kb.Build(),
		})

		if err != nil {
			m.Client.Log.Error(err.Error())
			_, _ = m.Reply("⚠️ Too many results. Please use a direct track URL or reduce playlist size (max 50).")
		}

		return nil
	}

	search, err := sp.Search("5")
	if err != nil {
		_, _ = m.Reply("❌ Failed to search for song: " + err.Error())
		return nil
	}

	if len(search.Results) == 0 {
		_, _ = m.Reply("😔 No results found.")
		return nil
	}

	for _, result := range search.Results {
		kb.AddRow(telegram.Button.Data(
			fmt.Sprintf("%s - %s", result.Name, result.Artist),
			fmt.Sprintf("spot_%s_%d", utils.EncodeURL(result.URL), m.SenderID()),
		))
	}

	_, err = m.Reply("<b>🎧 Select a song from below:</b>", telegram.SendOptions{
		ReplyMarkup: kb.Build(),
	})

	return err
}

// SpotifyHandlerCallback handles button presses for downloading Spotify songs.
func SpotifyHandlerCallback(cb *telegram.CallbackQuery) error {
	data := cb.DataString()
	first := strings.Index(data, "_")
	last := strings.LastIndex(data, "_")

	if first == -1 || last == -1 || first == last {
		_, _ = cb.Answer("❌ Invalid selection.", &telegram.CallbackOptions{Alert: true})
		_, _ = cb.Delete()
		return nil
	}

	vidID := data[first+1 : last]
	userIDStr := data[last+1:]
	if userIDStr != "0" && userIDStr != fmt.Sprint(cb.SenderID) {
		_, _ = cb.Answer("🚫 This action is not meant for you.", &telegram.CallbackOptions{Alert: true})
		return nil
	}

	_, _ = cb.Answer("🔄 Processing your request...", &telegram.CallbackOptions{Alert: true})

	url, err := utils.DecodeURL(vidID)
	if err != nil {
		cb.Client.Logger.Warn("Failed to decode URL:", err.Error())
		_, _ = cb.Edit("❌ Failed to decode the URL.")
		return nil
	}

	track, err := utils.NewApiData("").GetTrack(url)
	if err != nil {
		cb.Client.Logger.Warn("Failed to fetch track details:", err.Error())
		_, _ = cb.Edit("❌ Could not fetch track details.")
		return nil
	}

	msg, _ := cb.Edit("⏬ Downloading the song...")
	downloader := utils.NewDownload(*track)
	audioFile, thumbnail, err := downloader.Process()
	if err != nil || audioFile == "" {
		if err != nil {
			cb.Client.Logger.Warn("Failed to download or process the song:", err.Error())
		}
		_, _ = msg.Edit("⚠️ Failed to download the song.")
		return nil
	}

	progress := telegram.NewProgressManager(5)
	msg, _ = msg.Edit("⏫ Uploading the song...")

	re := regexp.MustCompile(`https?://t\.me/([^/]+)/(\d+)`)
	if matches := re.FindStringSubmatch(audioFile); len(matches) == 3 {
		username := matches[1]
		messageID, err := strconv.Atoi(matches[2])
		if err == nil {
			msgRef, err := msg.Client.GetMessageByID(username, int32(messageID))
			if err == nil {
				audioFile, err = msgRef.Download(&telegram.DownloadOptions{FileName: msgRef.File.Name})
				if err != nil {
					_, _ = msg.Edit("⚠️ Could not download the song file. " + err.Error())
					return nil
				}
			}
		}
	}

	progress.Edit(telegram.MediaDownloadProgress(msg, progress))

	caption := fmt.Sprintf("<b>🎶 %s - %d</b>\n<b>Artist:</b> %s", track.Name, track.Year, track.Artist)
	options := prepareTrackMessageOptions(track, audioFile, thumbnail, progress, caption)

	if _, err := msg.Edit(caption, options); err != nil {
		cb.Client.Logger.Warn("Failed to send track:", err.Error())
		_, _ = msg.Edit("❌ Failed to send the track.")
		return nil
	}

	_ = os.Remove(audioFile)
	return nil
}
