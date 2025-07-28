package src

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/amarnathcjd/gogram/telegram"
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
	args := m.Text()
	reqUrl := fmt.Sprintf("http://localhost:3010/snap?url=%s", args)
	m.Client.Log.Info("[FETCHING] " + reqUrl)

	resp, err := http.Get(reqUrl)
	if err != nil {
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

	var result APIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	// If videos exist
	if len(result.Video) > 0 {
		var album []string
		for _, v := range result.Video {
			album = append(album, v.Video)
		}
		_, err := m.ReplyAlbum(album, &telegram.MediaOptions{MimeType: "video/mp4"})
		if err != nil {
			return err
		}
	}

	// If images exist
	switch len(result.Image) {
	case 0:
		if len(result.Video) == 0 {
			m.Client.Log.Warn("No image or video found")
			return nil
		}
	case 1:
		_, err := m.ReplyMedia(result.Image[0], telegram.MediaOptions{FileName: "image.jpg", MimeType: "image/jpeg"})
		if err != nil {
			return err
		}
	default:
		var album []string
		for _, img := range result.Image {
			album = append(album, img)
		}
		_, err := m.ReplyAlbum(album)
		if err != nil {
			return err
		}
	}

	return nil
}
