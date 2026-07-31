package commands

import (
	"fmt"
	"html"
	"noinoi/internal/database"
	"noinoi/internal/httpx"
	"os"
	"strings"

	"github.com/AshokShau/gotdbot"
)

type SnapMediaItem struct {
	URL       string
	Thumbnail string
}

func snapHandler(c *gotdbot.Client, m *gotdbot.Message) error {
	targetUrl := getUrl(m)
	if targetUrl == "" {
		return nil
	}

	c.Logger.Info("Snap download request", "user_id", m.SenderID(), "chat_id", m.ChatId, "url", targetUrl)

	botId := c.Me.Id

	reply, err := m.ReplyText(c, "⏳ Processing...", nil)
	if err != nil {
		return err
	}

	snapData, err := httpx.GetSnap(targetUrl)
	if err != nil {
		database.IncrementDownloads(botId, false)
		_, _ = reply.EditText(c, fmt.Sprintf("Error: %v", err), nil)
		return nil
	}
	var caption string

	if m.IsPrivate() {
		rawCaption := snapData.Title
		if len(rawCaption) > 1000 {
			rawCaption = rawCaption[:1000] + "..."
		}
		caption = html.EscapeString(rawCaption)
	} else {
		caption = "Join @FallenProjects"
	}

	if len(snapData.Images) > 0 {
		for i := 0; i < len(snapData.Images); i += 10 {
			end := min(i+10, len(snapData.Images))

			batchUrls := snapData.Images[i:end]
			var batch []SnapMediaItem
			for _, u := range batchUrls {
				batch = append(batch, SnapMediaItem{URL: u})
			}
			if len(batch) == 1 {
				_, err = handleMediaUpload(c, m, batch[0], "photo", caption)
			} else {
				err = sendMediaAlbum(c, m, batch, "photo", caption)
			}

			if err != nil {
				database.IncrementDownloads(botId, false)
				_, _ = reply.EditText(c, fmt.Sprintf("Failed to send photo(s): %v", err), nil)
			} else {
				database.IncrementDownloads(botId, true)
			}
		}
	}

	if len(snapData.Audios) > 0 {
		var audioUrls []string
		for _, a := range snapData.Audios {
			if a.URL != "" {
				audioUrls = append(audioUrls, a.URL)
			}
		}

		for i := 0; i < len(audioUrls); i += 10 {
			end := min(i+10, len(audioUrls))
			batchUrls := audioUrls[i:end]
			var batch []SnapMediaItem
			for _, u := range batchUrls {
				batch = append(batch, SnapMediaItem{URL: u})
			}
			if len(batch) == 1 {
				_, err = handleMediaUpload(c, m, batch[0], "audio", caption)
			} else {
				err = sendMediaAlbum(c, m, batch, "audio", caption)
			}

			if err != nil {
				database.IncrementDownloads(botId, false)
			} else {
				database.IncrementDownloads(botId, true)
			}
		}
	}

	if len(snapData.Videos) > 0 {
		var videosWithAudio, videosWithoutAudio []SnapMediaItem
		for _, v := range snapData.Videos {
			if v.URL == "" {
				continue
			}
			hasAudio, _ := httpx.HasAudioStream(v.URL)
			if hasAudio {
				videosWithAudio = append(videosWithAudio, SnapMediaItem{URL: v.URL, Thumbnail: v.Thumbnail})
			} else {
				videosWithoutAudio = append(videosWithoutAudio, SnapMediaItem{URL: v.URL, Thumbnail: v.Thumbnail})
			}
		}

		for i := 0; i < len(videosWithAudio); i += 10 {
			end := min(i+10, len(videosWithAudio))
			batch := videosWithAudio[i:end]

			if len(batch) == 1 {
				_, err = handleMediaUpload(c, m, batch[0], "video", caption)
			} else {
				err = sendMediaAlbum(c, m, batch, "video", caption)
			}

			if err != nil {
				database.IncrementDownloads(botId, false)
			} else {
				database.IncrementDownloads(botId, true)
			}
		}

		for _, item := range videosWithoutAudio {
			_, err = handleMediaUpload(c, m, item, "animation", caption)
			if err != nil {
				database.IncrementDownloads(botId, false)
			} else {
				database.IncrementDownloads(botId, true)
			}
		}
	}

	_ = reply.Delete(c, true)
	return gotdbot.EndGroups
}

func handleMediaUpload(c *gotdbot.Client, m *gotdbot.Message, item SnapMediaItem, mediaType, caption string) (*gotdbot.Message, error) {
	var input gotdbot.InputFile
	input = &gotdbot.InputFileRemote{Id: item.URL}

	var err error
	var msg *gotdbot.Message

	opts := &gotdbot.SendPhotoOpts{
		Caption:   caption,
		ParseMode: "HTML",
	}

	switch mediaType {
	case "photo":
		msg, err = m.ReplyPhoto(c, input, opts)
	case "video":
		msg, err = m.ReplyVideo(c, input, &gotdbot.SendVideoOpts{Caption: caption, ParseMode: "HTML"})
	case "animation":
		msg, err = m.ReplyAnimation(c, input, &gotdbot.SendAnimationOpts{Caption: caption, ParseMode: "HTML"})
	case "audio":
		msg, err = m.ReplyAudio(c, input, &gotdbot.SendAudioOpts{Caption: caption, ParseMode: "HTML"})
	default:
		return nil, fmt.Errorf("unsupported media type: %s", mediaType)
	}
	if err != nil && (strings.Contains(err.Error(), "WEBPAGE_CURL_FAILED") || strings.Contains(err.Error(), "WEBPAGE_MEDIA_EMPTY")) {
		localPath, dlErr := httpx.DownloadFile(item.URL)
		if dlErr == nil {
			defer os.Remove(localPath)
			input = &gotdbot.InputFileLocal{Path: localPath}
			var thumb *gotdbot.InputThumbnail
			if item.Thumbnail != "" {
				thumbLocalPath, thumbDlErr := httpx.DownloadImg(item.Thumbnail)
				if thumbDlErr == nil {
					defer os.Remove(thumbLocalPath)
					thumb = &gotdbot.InputThumbnail{Thumbnail: &gotdbot.InputFileLocal{Path: thumbLocalPath}}
				}
			}

			switch mediaType {
			case "photo":
				msg, err = m.ReplyPhoto(c, input, opts)
			case "video":
				msg, err = m.ReplyVideo(c, input, &gotdbot.SendVideoOpts{Caption: caption, ParseMode: "HTML", Thumbnail: thumb})
			case "animation":
				msg, err = m.ReplyAnimation(c, input, &gotdbot.SendAnimationOpts{Caption: caption, ParseMode: "HTML", Thumbnail: thumb})
			case "audio":
				msg, err = m.ReplyAudio(c, input, &gotdbot.SendAudioOpts{Caption: caption, ParseMode: "HTML", AlbumCoverThumbnail: thumb})
			}
		}
	}

	return msg, err
}

func sendMediaAlbum(c *gotdbot.Client, m *gotdbot.Message, items []SnapMediaItem, mediaType, caption string) error {
	albumFunc := func(mediaItems []SnapMediaItem, isLocal bool) (*gotdbot.Messages, []string, error) {
		var contents []gotdbot.InputMessageContent
		var captionObj *gotdbot.FormattedText
		var err error
		var tempFiles []string

		if caption != "" {
			captionObj, err = c.ParseTextEntities(&gotdbot.TextParseModeHTML{}, caption)
			if err != nil {
				captionObj = &gotdbot.FormattedText{Text: "#FA"}
			}
		}

		for i, item := range mediaItems {
			var input gotdbot.InputFile
			if isLocal {
				input = &gotdbot.InputFileLocal{Path: item.URL}
			} else {
				input = &gotdbot.InputFileRemote{Id: item.URL}
			}

			var currentCaption *gotdbot.FormattedText
			if i == 0 {
				currentCaption = captionObj
			}

			var thumb *gotdbot.InputThumbnail
			if isLocal && item.Thumbnail != "" {
				thumbLocalPath, thumbDlErr := httpx.DownloadImg(item.Thumbnail)
				if thumbDlErr == nil {
					tempFiles = append(tempFiles, thumbLocalPath)
					thumb = &gotdbot.InputThumbnail{Thumbnail: &gotdbot.InputFileLocal{Path: thumbLocalPath}}
				}
			}

			switch mediaType {
			case "photo":
				contents = append(contents, &gotdbot.InputMessagePhoto{
					Photo: &gotdbot.InputPhoto{
						Photo: input,
					},
					Caption: currentCaption,
				})
			case "video":
				contents = append(contents, &gotdbot.InputMessageVideo{
					Video: &gotdbot.InputVideo{
						Video:     input,
						Thumbnail: thumb,
					},
					Caption: currentCaption,
				})
			case "animation":
				contents = append(contents, &gotdbot.InputMessageAnimation{
					Animation: &gotdbot.InputAnimation{
						Animation: input,
						Thumbnail: thumb,
					},
					Caption: currentCaption,
				})
			case "audio":
				contents = append(contents, &gotdbot.InputMessageAudio{
					Audio: &gotdbot.InputAudio{
						Audio:               input,
						AlbumCoverThumbnail: thumb,
					},
					Caption: currentCaption,
				})
			}
		}

		msg, err := c.SendMessageAlbum(m.ChatId, contents, &gotdbot.SendMessageAlbumOpts{
			ReplyTo: &gotdbot.InputMessageReplyToMessage{MessageId: m.Id},
		})

		return msg, tempFiles, err
	}

	_, _, err := albumFunc(items, false)
	if err != nil && (strings.Contains(err.Error(), "WEBPAGE_CURL_FAILED") || strings.Contains(err.Error(), "Group send failed") || strings.Contains(err.Error(), "WEBPAGE_MEDIA_EMPTY")) {
		var localItems []SnapMediaItem
		var mainPaths []string
		for _, item := range items {
			path, dlErr := httpx.DownloadFile(item.URL)
			if dlErr == nil {
				mainPaths = append(mainPaths, path)
				localItems = append(localItems, SnapMediaItem{URL: path, Thumbnail: item.Thumbnail})
			}
		}

		if len(localItems) > 0 {
			_, thumbPaths, newErr := albumFunc(localItems, true)
			err = newErr
			for _, p := range mainPaths {
				os.Remove(p)
			}
			for _, p := range thumbPaths {
				os.Remove(p)
			}
		}
	}

	return err
}
