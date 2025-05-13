package src

import (
	"context"
	"fmt"
	"github.com/amarnathcjd/gogram/telegram"
	"os"
	"songBot/config"
	"strings"
	"time"

	yt "github.com/lrstanley/go-ytdlp"
)

func YtVideoDL(m *telegram.NewMessage) error {
	yt.MustInstall(context.TODO(), nil)

	args := m.Args()
	if args == "" {
		m.Reply("Provide video url~")
		return nil
	}

	msg, _ := m.Reply("Downloading video...")

	// Constants for progress reporting
	const (
		progressUpdateInterval = 7 * time.Second
		progressBarLength      = 10
	)

	dl := yt.New().
		//FormatSort("res:1080,tbr").
		Format("(bestvideo[height<=?720][width<=?1280][ext=mp4])+(bestaudio[ext=m4a])").
		NoWarnings().
		RecodeVideo("mp4").
		Output("yt-video.mp4").
		ProgressFunc(progressUpdateInterval, func(update yt.ProgressUpdate) {
			// Template for progress message
			text := "<b>~ Downloading Youtube Video ~</b>\n\n" +
				"<b>📄 Name:</b> <code>%s</code>\n" +
				"<b>💾 File Size:</b> <code>%.2f MiB</code>\n" +
				"<b>⌛️ ETA:</b> <code>%s</code>\n" +
				"<b>⏱ Speed:</b> <code>%s</code>\n" +
				"<b>⚙️ Progress:</b> %s <code>%.2f%%</code>"

			// Calculate file size in MiB
			size := float64(update.TotalBytes) / (1024 * 1024)

			// Calculate ETA
			eta := calculateETA(update)

			// Calculate download speed
			speed := calculateSpeed(update)

			// Calculate download percentage
			percent := float64(update.DownloadedBytes) / float64(update.TotalBytes) * 100
			if percent == 0 {
				msg.Edit("Starting download...")
				return
			}

			// Create progress bar
			progressbar := createProgressBar(percent, progressBarLength)

			message := fmt.Sprintf(text, *update.Info.Title, size, eta, speed, progressbar, percent)
			msg.Edit(message)
		}).
		Proxy(config.Proxy).
		NoWarnings()

	_, err := dl.Run(context.TODO(), args)
	if err != nil {
		msg.Edit("<code>video not found.</code>")
		return nil
	}

	defer os.Remove("yt-video.mp4")
	defer msg.Delete()

	m.ReplyMedia("yt-video.mp4", telegram.MediaOptions{
		Attributes: []telegram.DocumentAttribute{
			&telegram.DocumentAttributeFilename{
				FileName: "yt-video.mp4",
			},
		},
		ProgressManager: telegram.NewProgressManager(5).SetMessage(msg),
	})
	return nil
}

// calculateETA computes the estimated time remaining for download completion
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

// calculateSpeed computes the current download speed
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

// createProgressBar generates a visual progress bar
func createProgressBar(percent float64, length int) string {
	filled := int(percent / 100 * float64(length))
	empty := length - filled
	return strings.Repeat("■", filled) + strings.Repeat("□", empty)
}

// formatDuration formats a duration in a human-readable way
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
