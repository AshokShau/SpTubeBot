package src

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/amarnathcjd/gogram/telegram"
)

type YTData struct {
	Success bool `json:"success"`
	Data    struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Duration int    `json:"duration"`
		Author   string `json:"author"`
		Image    string `json:"image"`
		Videos   []struct {
			URL      string `json:"url"`
			Quality  string `json:"quality"`
			Filesize int    `json:"filesize"`
		} `json:"videos"`
		Audios []struct {
			URL      string  `json:"url"`
			Quality  float64 `json:"quality"`
			Filesize int     `json:"filesize"`
		} `json:"audios"`
	} `json:"data"`
}

func isURL(s string) bool {
	re := regexp.MustCompile(`^https?://(www\.)?(youtube\.com|youtu\.be|music\.youtube\.com)/[^\s]+$`)
	return re.MatchString(s)
}

func formatDuration(seconds int) string {
	_min := seconds / 60
	sec := seconds % 60
	return fmt.Sprintf("%02d:%02d", _min, sec)
}

func formatSize(bytes int) string {
	return fmt.Sprintf("%.2f MB", float64(bytes)/1024/1024)
}

func YtVideoDL(m *telegram.NewMessage) error {
	args := m.Args()
	if args == "" {
		_, _ = m.Reply("🚫 Please send a valid YouTube URL after the command.")
		return nil
	}

	if !isURL(args) {
		_, _ = m.Reply("❌ That doesn't look like a valid URL. Please check and try again.")
		return nil
	}

	api := "https://info.fallenapi.fun/youtube/info?url=" + url.QueryEscape(args)
	resp, err := http.Get(api)
	if err != nil || resp.StatusCode != 200 {
		_, _ = m.Reply("⚠️ Failed to fetch video info. Please try again later.")
		return nil
	}
	defer resp.Body.Close()

	var data YTData
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil || !data.Success {
		_, _ = m.Reply("😞 Could not parse video data. Something went wrong.")
		return nil
	}

	title := data.Data.Title
	image := data.Data.Image
	author := data.Data.Author
	duration := formatDuration(data.Data.Duration)

	videoURL := data.Data.Videos[0].URL
	videoQuality := data.Data.Videos[0].Quality
	videoSize := formatSize(data.Data.Videos[0].Filesize)

	audioURL := data.Data.Audios[0].URL
	audioQuality := fmt.Sprintf("%.1f kbps", data.Data.Audios[0].Quality)
	audioSize := formatSize(data.Data.Audios[0].Filesize)

	caption := fmt.Sprintf(
		`🎬 <b>%s</b>

👤 <b>Channel:</b> %s
⏱️ <b>Duration:</b> %s

⬇️ <b>Click the buttons below to download:</b>

🎥 <b>Video:</b> %s • %s
🎧 <b>Audio:</b> %s • %s`,
		title, author, duration,
		videoQuality, videoSize,
		audioQuality, audioSize,
	)

	opts := telegram.MediaOptions{
		ReplyMarkup: telegram.NewKeyboard().
			AddRow(
				telegram.Button.URL("🎥 Download Video", videoURL),
				telegram.Button.URL("🎧 Download Audio", audioURL),
			).
			AddRow(
				telegram.Button.URL("💫 Fᴀʟʟᴇɴ Pʀᴏᴊᴇᴄᴛs", "https://t.me/FallenProjects"),
			).Build(),
		Caption: caption,
	}
	
	_, err = m.ReplyMedia(image, opts)
	if err != nil {
		return err
	}
	return nil
}
