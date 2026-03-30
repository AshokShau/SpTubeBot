package commands

import (
	"fmt"
	"html"
	"noinoi/internal/httpx"
	"os"
	"strings"

	"github.com/AshokShau/gotdbot"
)

func snapHandler(c *gotdbot.Client, ctx *gotdbot.Context) error {
	m := ctx.EffectiveMessage
	targetUrl := getUrl(m)
	if targetUrl == "" {
		return nil
	}

	reply, err := m.ReplyText(c, "⏳ Processing...", nil)
	if err != nil {
		return err
	}

	snapData, err := httpx.GetSnap(targetUrl)
	if err != nil {
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
			end := i + 10
			if end > len(snapData.Images) {
				end = len(snapData.Images)
			}

			batch := snapData.Images[i:end]
			if len(batch) == 1 {
				_, err = handleMediaUpload(c, m, batch[0], "photo", caption)
			} else {
				err = sendMediaAlbum(c, m, batch, "photo", caption)
			}

			if err != nil {
				_, _ = reply.EditText(c, fmt.Sprintf("Failed to send photo(s): %v", err), nil)
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
			end := i + 10
			if end > len(audioUrls) {
				end = len(audioUrls)
			}
			batch := audioUrls[i:end]

			if len(batch) == 1 {
				_, _ = handleMediaUpload(c, m, batch[0], "audio", caption)
			} else {
				_ = sendMediaAlbum(c, m, batch, "audio", caption)
			}
		}
	}

	if len(snapData.Videos) > 0 {
		var videosWithAudio, videosWithoutAudio []string
		for _, v := range snapData.Videos {
			if v.URL == "" {
				continue
			}
			hasAudio, _ := httpx.HasAudioStream(v.URL)
			if hasAudio {
				videosWithAudio = append(videosWithAudio, v.URL)
			} else {
				videosWithoutAudio = append(videosWithoutAudio, v.URL)
			}
		}

		for i := 0; i < len(videosWithAudio); i += 10 {
			end := i + 10
			if end > len(videosWithAudio) {
				end = len(videosWithAudio)
			}
			batch := videosWithAudio[i:end]

			if len(batch) == 1 {
				_, _ = handleMediaUpload(c, m, batch[0], "video", caption)
			} else {
				_ = sendMediaAlbum(c, m, batch, "video", caption)
			}
		}

		for _, url := range videosWithoutAudio {
			_, _ = handleMediaUpload(c, m, url, "animation", caption)
		}
	}

	_ = reply.Delete(c, true)
	return gotdbot.EndGroups
}

func handleMediaUpload(c *gotdbot.Client, m *gotdbot.Message, mediaUrl, mediaType, caption string) (*gotdbot.Message, error) {
	var input gotdbot.InputFile
	input = &gotdbot.InputFileRemote{Id: mediaUrl}

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
		ext := ".jpg"
		if mediaType == "video" || mediaType == "animation" {
			ext = ".mp4"
		} else if mediaType == "audio" {
			ext = ".mp3"
		}

		localPath, dlErr := httpx.DownloadFileToTemp(mediaUrl, ext)
		if dlErr == nil {
			defer os.Remove(localPath)
			input = &gotdbot.InputFileLocal{Path: localPath}
			switch mediaType {
			case "photo":
				msg, err = m.ReplyPhoto(c, input, opts)
			case "video":
				msg, err = m.ReplyVideo(c, input, &gotdbot.SendVideoOpts{Caption: caption, ParseMode: "HTML"})
			case "animation":
				msg, err = m.ReplyAnimation(c, input, &gotdbot.SendAnimationOpts{Caption: caption, ParseMode: "HTML"})
			case "audio":
				msg, err = m.ReplyAudio(c, input, &gotdbot.SendAudioOpts{Caption: caption, ParseMode: "HTML"})
			}
		}
	}

	return msg, err
}

func sendMediaAlbum(c *gotdbot.Client, m *gotdbot.Message, mediaUrls []string, mediaType, caption string) error {
	albumFunc := func(urls []string) (*gotdbot.Messages, error) {
		var contents []gotdbot.InputMessageContent
		var captionObj *gotdbot.FormattedText
		var err error

		if caption != "" {
			captionObj, err = c.ParseTextEntities(&gotdbot.TextParseModeHTML{}, caption)
			if err != nil {
				captionObj = &gotdbot.FormattedText{Text: "#FA"}
			}
		}

		for i, url := range urls {
			var input gotdbot.InputFile
			if strings.HasPrefix(url, "http") {
				input = &gotdbot.InputFileRemote{Id: url}
			} else {
				input = &gotdbot.InputFileLocal{Path: url}
			}

			var currentCaption *gotdbot.FormattedText
			if i == 0 {
				currentCaption = captionObj
			}

			switch mediaType {
			case "photo":
				contents = append(contents, &gotdbot.InputMessagePhoto{Photo: input, Caption: currentCaption})
			case "video":
				contents = append(contents, &gotdbot.InputMessageVideo{Video: input, Caption: currentCaption})
			case "animation":
				contents = append(contents, &gotdbot.InputMessageAnimation{Animation: input, Caption: currentCaption})
			case "audio":
				contents = append(contents, &gotdbot.InputMessageAudio{Audio: input, Caption: currentCaption})
			}
		}

		msg, err := c.SendMessageAlbum(m.ChatId, contents, &gotdbot.SendMessageAlbumOpts{
			ReplyTo: &gotdbot.InputMessageReplyToMessage{MessageId: m.Id},
		})

		return msg, err
	}

	_, err := albumFunc(mediaUrls)
	if err != nil && strings.Contains(err.Error(), "WEBPAGE_CURL_FAILED") || err != nil && strings.Contains(err.Error(), "Group send failed") || err != nil && strings.Contains(err.Error(), "WEBPAGE_MEDIA_EMPTY") {
		ext := ".jpg"
		if mediaType == "video" || mediaType == "animation" {
			ext = ".mp4"
		} else if mediaType == "audio" {
			ext = ".mp3"
		}

		var localPaths []string
		for _, url := range mediaUrls {
			path, dlErr := httpx.DownloadFileToTemp(url, ext)
			if dlErr == nil {
				localPaths = append(localPaths, path)
			}
		}

		if len(localPaths) > 0 {
			defer func() {
				for _, p := range localPaths {
					os.Remove(p)
				}
			}()
			_, err = albumFunc(localPaths)
		}
	}

	return err
}
