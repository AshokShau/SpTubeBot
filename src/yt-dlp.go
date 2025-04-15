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

	dl := yt.New().
		FormatSort("res,ext:mp4:m4a").
		Format("bv+ba").
		RecodeVideo("mp4").
		Output("yt-video.mp4").
		ProgressFunc(time.Second*7, func(update yt.ProgressUpdate) {
			text := "<b>~ Downloading Youtube Video ~</b>\n\n"
			text += "<b>📄 Name:</b> <code>%s</code>\n"
			text += "<b>💾 File Size:</b> <code>%.2f MiB</code>\n"
			text += "<b>⌛️ ETA:</b> <code>%s</code>\n"
			text += "<b>⏱ Speed:</b> <code>%s</code>\n"
			text += "<b>⚙️ Progress:</b> %s <code>%.2f%%</code>"

			size := float64(update.TotalBytes) / 1024 / 1024
			eta := func() string {
				elapsed := time.Now().Unix() - update.Started.Unix()
				remaining := float64(update.TotalBytes-update.DownloadedBytes) / float64(update.DownloadedBytes) * float64(elapsed)
				return (time.Second * time.Duration(remaining)).String()
			}()

			speed := func() string {
				elapsedTime := time.Since(time.Unix(update.Started.Unix(), 0))
				if int(elapsedTime.Seconds()) == 0 {
					return "0 B/s"
				}
				speedBps := float64(update.TotalBytes) / elapsedTime.Seconds()
				if speedBps < 1024 {
					return fmt.Sprintf("%.2f B/s", speedBps)
				} else if speedBps < 1024*1024 {
					return fmt.Sprintf("%.2f KB/s", speedBps/1024)
				} else {
					return fmt.Sprintf("%.2f MB/s", speedBps/1024/1024)
				}
			}()
			percent := float64(update.DownloadedBytes) / float64(update.TotalBytes) * 100

			progressbar := strings.Repeat("■", int(percent/10)) + strings.Repeat("□", 10-int(percent/10))

			message := fmt.Sprintf(text, *update.Info.Title, size, eta, speed, progressbar, percent)
			msg.Edit(message)
		}).
		Proxy(config.Proxy).
		NoWarnings()

	_, err := dl.Run(context.TODO(), args)
	if err != nil {
		m.Edit("<code>video not found.</code>")
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
