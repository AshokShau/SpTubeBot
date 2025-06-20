package src

import (
	"fmt"
	"github.com/amarnathcjd/gogram/telegram"
	"os"
	"regexp"
	"songBot/src/utils"
	"strconv"
	"strings"
)

// SpotifySearchSong handles song search requests from messages.
func SpotifySearchSong(m *telegram.NewMessage) error {
	songName := m.Text()
	if m.IsCommand() {
		songName = m.Args()
	}

	if songName == "" {
		_, err := m.Reply("Please provide a song name to search. or spotify url.")
		return err
	}

	sp := utils.NewApiData(songName)
	kb := telegram.NewKeyboard()
	if sp.IsValid(songName) {
		song, err := sp.GetInfo(songName)
		if err != nil {
			_, _ = m.Reply(fmt.Sprintf("Error: %v", err))
			return nil
		}

		if song == nil || len(song.Results) == 0 {
			_, _ = m.Reply("Song not found.")
			return nil
		}

		for _, track := range song.Results {
			data := fmt.Sprintf("spot_%s_0", utils.EncodeURL(track.URL))
			kb.AddRow(telegram.Button.Data(
				fmt.Sprintf("%s - %s", track.Name, track.Artist),
				data,
			))
		}

		_, err = m.Reply("<b>Select a song from below:</b>", telegram.SendOptions{
			ReplyMarkup: kb.Build(),
		})

		if err != nil {
			m.Client.Log.Error(err.Error())
			_, _ = m.Reply("Too many results. plz use track url. or less then 50 songs in playlist.")
			return nil
		}

		return nil
	}

	search, err := sp.Search("5")
	if err != nil {
		_, _ = m.Reply("Failed to search for song." + err.Error())
		return nil
	}

	if len(search.Results) == 0 {
		_, _ = m.Reply("No results found.")
		return nil
	}

	for _, result := range search.Results {
		kb.AddRow(telegram.Button.Data(
			fmt.Sprintf("%s - %s", result.Name, result.Artist),
			fmt.Sprintf("spot_%s_%d", utils.EncodeURL(result.URL), m.SenderID()),
		))
	}

	// Send the search results with a keyboard
	_, err = m.Reply("<b>Select a song from below:</b>", telegram.SendOptions{
		ReplyMarkup: kb.Build(),
	})

	if err != nil {
		return err
	}

	return nil
}

func SpotifyHandlerCallback(cb *telegram.CallbackQuery) error {
	data := cb.DataString()
	firstUnderscore := strings.Index(data, "_")
	lastUnderscore := strings.LastIndex(data, "_")
	if firstUnderscore == -1 || lastUnderscore == -1 || firstUnderscore == lastUnderscore {
		_, _ = cb.Answer("Invalid selection.", &telegram.CallbackOptions{Alert: true})
		_, _ = cb.Delete()
		return nil
	}
	vidID := data[firstUnderscore+1 : lastUnderscore]
	userId := data[lastUnderscore+1:]
	userID := fmt.Sprintf("%d", cb.SenderID)
	if userId != "0" && userId != userID {
		_, _ = cb.Answer("This action is not intended for you.", &telegram.CallbackOptions{Alert: true})
		return nil
	}

	_, _ = cb.Answer("Processing your request...", &telegram.CallbackOptions{Alert: true})

	url, err := utils.DecodeURL(vidID)
	if err != nil {
		cb.Client.Logger.Warn("Failed to decode URL:", err.Error())
		_, _ = cb.Edit("Failed to decode URL. Please try again later.")
		return nil
	}

	track, err := utils.NewApiData("").GetTrack(url)
	if err != nil {
		cb.Client.Logger.Warn("Failed to fetch track details: " + err.Error())
		_, _ = cb.Edit("Failed to fetch track details. Please try again later.")
		return nil
	}

	message, _ := cb.Edit("Downloading the song...")
	downloader := utils.NewDownload(*track)
	audioFile, thumbnail, err := downloader.Process()
	if err != nil {
		cb.Client.Logger.Warn("Failed to process the song:", err.Error())
		_, _ = message.Edit("Failed to process the song. Please try again later.")
		return nil
	}

	if audioFile == "" {
		cb.Client.Logger.Warn("Failed to download the song:")
		_, _ = message.Edit("Failed to download the song. Please try again later.")
		return nil
	}
	progressManager := telegram.NewProgressManager(5)
	message, _ = message.Edit("Uploading the song...")
	re := regexp.MustCompile(`https?://t\.me/([^/]+)/(\d+)`)

	matches := re.FindStringSubmatch(audioFile)
	if len(matches) == 3 {
		username := matches[1]
		messageID, err := strconv.Atoi(matches[2])
		if err != nil {
			cb.Client.Logger.Warn("Failed to upload the song:", err.Error())
			_, err = message.Edit("Failed to upload the song. Please try again later.")
		}

		msgX, err := message.Client.GetMessageByID(username, int32(messageID))
		if err != nil {
			cb.Client.Logger.Warn("Failed to upload the song:", err.Error())
			_, err = message.Edit("Failed to upload the song. Please try again later.")
		}

		if audioFile, err = msgX.Download(&telegram.DownloadOptions{
			FileName: msgX.File.Name,
		}); err != nil {
			_, err = message.Edit("Failed to download the song. Please try again later." + err.Error())
			return nil
		}
	}

	progressManager.Edit(telegram.MediaDownloadProgress(message, progressManager))
	caption := fmt.Sprintf("<b> %s- %d</b>\n<b>Artist:</b> %s", track.Name, track.Year, track.Artist)
	options := prepareTrackMessageOptions(track, audioFile, thumbnail, progressManager, caption)
	_, err = message.Edit(caption, options)
	if err != nil {
		cb.Client.Logger.Warn("Failed to send the track:", err.Error())
		_, _ = message.Edit("Failed to send the track.")
		return nil
	}

	_ = os.Remove(audioFile)
	return nil
}
