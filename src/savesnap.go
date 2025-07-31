package src

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/amarnathcjd/gogram/telegram"
	"io"
	"net/http"
	"net/url"
	"os"
	"songBot/src/config"
	"songBot/src/utils"
	"strings"
	"time"
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
	rawURL := m.Text()
	endpoint := fmt.Sprintf("%s/snap?url=%s", config.ApiUrl, url.QueryEscape(rawURL))
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("creating request failed: %v", err)
	}

	req.Header.Set("X-API-Key", config.ApiKey)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
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
		_, sendErr = m.ReplyAlbum(data.Image, &telegram.MediaOptions{
			MimeType: "image/jpeg",
		})
		if sendErr != nil && strings.Contains(sendErr.Error(), "MEDIA_EMPTY") {
			filePaths := make([]string, 0, len(data.Image))
			for _, imgUrl := range data.Image {
				randomBytes := make([]byte, 4)
				_, _ = rand.Read(randomBytes)
				fileName := hex.EncodeToString(randomBytes) + ".jpg"

				filePath, err := utils.DownloadFile(context.Background(), imgUrl, fileName, true)
				if err != nil {
					continue // skip failed downloads
				}
				filePaths = append(filePaths, filePath)
			}

			if len(filePaths) > 0 {
				_, sendErr = m.ReplyAlbum(filePaths, &telegram.MediaOptions{MimeType: "image/jpeg"})
			}

			defer func() {
				for _, filePath := range filePaths {
					_ = os.Remove(filePath)
				}
			}()
		}
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

	if sendErr != nil {
		_, _ = m.Reply("Error: " + sendErr.Error())
	}
	return sendErr
}
