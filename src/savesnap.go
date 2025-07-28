package src

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/amarnathcjd/gogram/telegram"
	"io"
	"net/http"
	"os"
	"songBot/src/utils"
	"strings"
)

type APIVideo struct {
	Video     string `json:"video"`
	Thumbnail string `json:"thumbnail"`
}

type APIResponse struct {
	Video []APIVideo `json:"video"`
	Image []string   `json:"image"`
	Fetch bool       `json:"fetch"`
}

func saveSnap(m *telegram.NewMessage) error {
	url := m.Text()
	apiURL := fmt.Sprintf("https://info.fallenapi.fun/snap?url=%s", url)
	m.Client.Log.Info("[FETCHING] " + apiURL)

	resp, err := http.Get(apiURL)
	if err != nil {
		m.Client.Log.Error("HTTP error: " + err.Error())
		return fmt.Errorf("failed to fetch snap: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var data APIResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	var sendErr error
	// Handle images
	switch len(data.Image) {
	case 0:
		// No image, skip
	case 1:
		_, sendErr = m.ReplyMedia(data.Image[0], telegram.MediaOptions{
			FileName: "image.jpg",
			MimeType: "image/jpeg",
		})
	default:
		var images []string
		for _, img := range data.Image {
			images = append(images, img)
		}
		_, sendErr = m.ReplyAlbum(images, &telegram.MediaOptions{
			MimeType: "image/jpeg",
		})
	}

	// Handle videos
	switch len(data.Video) {
	case 0:
		// No video, skip
	case 1:
		v := data.Video[0]
		_, sendErr = m.ReplyMedia(v.Video, telegram.MediaOptions{
			FileName: "video.mp4",
			MimeType: "video/mp4",
		})
		if sendErr != nil && strings.Contains(sendErr.Error(), "WEBPAGE_CURL_FAILED") {
			msg, err := m.Reply("Downloading...")
			if err != nil {
				return err
			}

			filePath, err := utils.DownloadFile(context.Background(), v.Video, "", false)
			if err != nil {
				_, err = msg.Edit("Download failed: " + err.Error())
				return err
			}
			progress := telegram.NewProgressManager(4)
			progress.Edit(telegram.MediaDownloadProgress(msg, progress))
			defer func() {
				_ = os.Remove(filePath)
			}()

			opts := telegram.SendOptions{
				ProgressManager: progress,
				Media:           filePath,
				Caption:         "Done!",
				MimeType:        "",
			}
			_, sendErr = msg.Edit("Done!", opts)
		}
	default:
		var videos []string
		for _, v := range data.Video {
			videos = append(videos, v.Video)
		}
		_, sendErr = m.ReplyAlbum(videos, &telegram.MediaOptions{
			MimeType: "video/mp4",
		})
	}

	return sendErr
}
