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

	// Check if file is a Telegram link (e.g., https://t.me/channel/1234)
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

	if !fileExists(audioFile) {
		cb.Client.Logger.Warn("Audio file does not exist:", audioFile)
		_, _ = msg.Edit("❌ Audio file missing.")
		return nil
	}

	progress := telegram.NewProgressManager(4)
	progress.Edit(telegram.MediaDownloadProgress(msg, progress))
	opts := prepareTrackMessageOptions(audioFile, thumb, track, progress)
	_, err = msg.Edit(buildTrackCaption(track), opts)

	if err != nil {
		_, _ = msg.Edit("❌ Failed to send the track. " + err.Error())
		return nil
	}

	cb.Client.Logger.Debug("Successfully sent track.")
	return nil
}

func zipHandle(m *telegram.NewMessage) error {
	query := strings.TrimSpace(m.Args())
	if query == "" {
		_, err := m.Reply("🎵 Please send me a song name, artist, or Spotify URL.\nExample: /playlist Daft Punk Get Lucky")
		return err
	}

	api := utils.NewApiData(query)
	var tracks *utils.PlatformTracks
	var err error
	msg, err := m.Reply("🔍 Searching for tracks...")
	if err != nil {
		return nil
	}

	if !api.IsValid(query) {
		tracks, err = api.Search("5")
	} else {
		tracks, err = api.GetInfo()
	}

	if err != nil || len(tracks.Results) == 0 {
		_, _ = msg.Edit("⚠️ Couldn't find any tracks. Please try a different search.")
		return nil
	}

	if tracks.Results[0].Platform == "youtube" {
		_, _ = msg.Edit("⚠️ YouTube is not supported. Please try a different search.")
		return nil
	}

	msg, _ = msg.Edit(fmt.Sprintf("⏳ Found %d tracks. Preparing download...", len(tracks.Results)))

	// Create ZIP file
	zipResult, err := utils.ZipTracks(tracks)
	if err != nil {
		_, _ = msg.Edit("❌ Failed to create zip file. Please try again later." + err.Error())
		return nil
	}

	if !fileExists(zipResult.ZipPath) {
		_, _ = msg.Edit("⚠️ Download completed but zip file is missing. Please report this issue.")
		return nil
	}

	// Prepare final message
	successMsg := fmt.Sprintf("✅ Success! Downloaded %d/%d tracks.\n📦 Zip file ready:",
		zipResult.SuccessCount,
		len(tracks.Results))

	if len(zipResult.Errors) > 0 {
		successMsg += fmt.Sprintf("\n\n⚠️ %d tracks failed to download.", len(zipResult.Errors))
	}

	_, err = msg.Edit(
		successMsg,
		telegram.SendOptions{
			Media:    zipResult.ZipPath,
			MimeType: "application/zip",
			Caption:  fmt.Sprintf("🎵 %d tracks", zipResult.SuccessCount),
		},
	)

	defer func() {
		_ = os.Remove(zipResult.ZipPath)
	}()

	if err != nil {
		_, _ = msg.Edit("❌ Failed to send zip file. Please try again later." + err.Error())
		return nil
	}

	return nil
}
