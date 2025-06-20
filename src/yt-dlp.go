package src

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/amarnathcjd/gogram/telegram"
	"io"
	"log"
	"net/http"
	"os"
	"songBot/src/config"
	"songBot/src/utils"
	"strconv"
	"strings"
	"time"

	yt "github.com/lrstanley/go-ytdlp"
)

func YtVideoDL(m *telegram.NewMessage) error {
	args := m.Args()
	if args == "" {
		m.Reply("Provide video URL~")
		return nil
	}

	msg, _ := m.Reply("Trying direct download from API...")

	apiUrl := fmt.Sprintf(
		"https://tgmusic.fallenapi.fun/yt?api_key=%s&id=%s&video=true",
		config.ApiKey,
		args,
	)
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Get(apiUrl)
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()

		var data struct {
			Results string `json:"results"`
		}

		body, _ := io.ReadAll(resp.Body)
		_ = json.Unmarshal(body, &data)

		if data.Results != "" {
			filePath, err := utils.DownloadFile(context.Background(), data.Results, "", false)
			if err == nil {
				defer os.Remove(filePath)
				defer msg.Delete()
				m.ReplyMedia(filePath, telegram.MediaOptions{
					Attributes: []telegram.DocumentAttribute{
						//&telegram.DocumentAttributeFilename{FileName: "yt-video.mp4"},
					},
					ProgressManager: telegram.NewProgressManager(5).SetMessage(msg),
				})
				return nil
			} else {
				msg.Edit("API download failed, falling back to yt-dlp...")
			}
		} else {
			msg.Edit("No direct download found. Using yt-dlp...")
		}
	} else {
		log.Println(err)
		msg.Edit("API unreachable. Using yt-dlp...")
	}

	yt.MustInstall(context.TODO(), nil)

	const (
		progressUpdateInterval = 7 * time.Second
		progressBarLength      = 10
	)

	dl := yt.New().
		Format("(bestvideo[height<=?720][width<=?1280][ext=mp4])+(bestaudio[ext=m4a])").
		NoWarnings().
		RecodeVideo("mp4").
		Output("yt-video.mp4").
		ProgressFunc(progressUpdateInterval, func(update yt.ProgressUpdate) {
			text := "<b>~ Downloading Youtube Video ~</b>\n\n" +
				"<b>📄 Name:</b> <code>%s</code>\n" +
				"<b>💾 File Size:</b> <code>%.2f MiB</code>\n" +
				"<b>⌛️ ETA:</b> <code>%s</code>\n" +
				"<b>⏱ Speed:</b> <code>%s</code>\n" +
				"<b>⚙️ Progress:</b> %s <code>%.2f%%</code>"

			size := float64(update.TotalBytes) / (1024 * 1024)
			eta := calculateETA(update)
			speed := calculateSpeed(update)
			percent := float64(update.DownloadedBytes) / float64(update.TotalBytes) * 100

			if percent == 0 {
				msg.Edit("Starting download...")
				return
			}

			progressbar := createProgressBar(percent, progressBarLength)
			message := fmt.Sprintf(text, *update.Info.Title, size, eta, speed, progressbar, percent)
			msg.Edit(message)
		}).
		Proxy(config.Proxy).
		NoWarnings().
		NoPart().
		Continue().
		Retries(strconv.Itoa(2))

	_, err = dl.Run(context.TODO(), args)
	if err != nil {
		_, _ = msg.Edit("<code>video not found.</code>")
		return nil
	}

	defer os.Remove("yt-video.mp4")
	defer msg.Delete()

	_, err = m.ReplyMedia("yt-video.mp4", telegram.MediaOptions{
		Attributes: []telegram.DocumentAttribute{
			//&telegram.DocumentAttributeFilename{FileName: "yt-video.mp4"},
		},
		ProgressManager: telegram.NewProgressManager(5).SetMessage(msg),
	})
	if err != nil {
		_, _ = msg.Edit("Error: " + err.Error())
		return err
	}
	return nil
}

func calculateETA(update yt.ProgressUpdate) string {
	if update.DownloadedBytes == 0 {
		return "calculating..."
	}

	elapsed := time.Since(update.Started)
	remainingBytes := update.TotalBytes - update.DownloadedBytes
	bytesPerSec := float64(update.DownloadedBytes) / elapsed.Seconds()

	if bytesPerSec <= 0 {
		return "unknown"
	}

	remainingTime := time.Duration(float64(remainingBytes)/bytesPerSec) * time.Second
	return formatDuration(remainingTime)
}

func calculateSpeed(update yt.ProgressUpdate) string {
	elapsed := time.Since(update.Started)
	if elapsed.Seconds() == 0 {
		return "0 B/s"
	}

	speedBps := float64(update.DownloadedBytes) / elapsed.Seconds()

	switch {
	case speedBps >= 1024*1024:
		return fmt.Sprintf("%.2f MB/s", speedBps/(1024*1024))
	case speedBps >= 1024:
		return fmt.Sprintf("%.2f KB/s", speedBps/1024)
	default:
		return fmt.Sprintf("%.2f B/s", speedBps)
	}
}

func createProgressBar(percent float64, length int) string {
	filled := int(percent / 100 * float64(length))
	empty := length - filled
	return strings.Repeat("■", filled) + strings.Repeat("□", empty)
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	switch {
	case h > 0:
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	default:
		return fmt.Sprintf("%02d:%02d", m, s)
	}
}
