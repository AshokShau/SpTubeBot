package commands

import (
	"fmt"
	"html"
	"noinoi/internal/httpx"
	"os"
	"strings"

	"github.com/AshokShau/gotdbot"
)

func musicHandler(c *gotdbot.Client, ctx *gotdbot.Context) error {
	m := ctx.EffectiveMessage
	targetUrl := getUrl(m)
	if targetUrl == "" {
		return nil
	}

	reply, err := m.ReplyText(c, "⏳ Processing Music...", nil)
	if err != nil {
		return err
	}

	musicInfo, err := httpx.GetMusicInfo(targetUrl)
	if err != nil {
		_, _ = reply.EditText(c, fmt.Sprintf("Error: %v", err), nil)
		return nil
	}

	if len(musicInfo.Results) == 0 {
		_, _ = reply.EditText(c, "No tracks found for this URL.", nil)
		return nil
	}

	// Only download the first track
	track := musicInfo.Results[0]

	trackDetails, err := httpx.GetTrack(track.URL)
	if err != nil {
		_, _ = reply.EditText(c, fmt.Sprintf("Error fetching track details: %v", err), nil)
		return nil
	}

	escapedTitle := html.EscapeString(track.Title)
	caption := fmt.Sprintf("<b>%s</b>\n\nJoin @FallenProjects", escapedTitle)

	var thumbInput *gotdbot.InputThumbnail
	if track.Thumbnail != "" {
		thumbInput = &gotdbot.InputThumbnail{Thumbnail: &gotdbot.InputFileRemote{Id: track.Thumbnail}}
	}

	input := &gotdbot.InputFileRemote{Id: trackDetails.CdnURL}

	_, err = m.ReplyAudio(c, input, &gotdbot.SendAudioOpts{
		Caption:             caption,
		ParseMode:           "HTML",
		Title:               track.Title,
		Performer:           track.Channel,
		Duration:            int32(track.Duration),
		AlbumCoverThumbnail: thumbInput,
	})

	if err != nil && (strings.Contains(err.Error(), "WEBPAGE_CURL_FAILED") || strings.Contains(err.Error(), "WEBPAGE_MEDIA_EMPTY")) {
		localPath, dlErr := httpx.DownloadFileToTemp(trackDetails.CdnURL, ".mp3")
		if dlErr == nil {
			defer os.Remove(localPath)
			localInput := &gotdbot.InputFileLocal{Path: localPath}

			var localThumbInput *gotdbot.InputThumbnail
			if track.Thumbnail != "" {
				thumbPath, thumbDlErr := httpx.DownloadFileToTemp(track.Thumbnail, ".jpg")
				if thumbDlErr == nil {
					defer os.Remove(thumbPath)
					localThumbInput = &gotdbot.InputThumbnail{Thumbnail: &gotdbot.InputFileLocal{Path: thumbPath}}
				}
			}

			_, err = m.ReplyAudio(c, localInput, &gotdbot.SendAudioOpts{
				Caption:             caption,
				ParseMode:           "HTML",
				Title:               track.Title,
				Performer:           track.Channel,
				Duration:            int32(track.Duration),
				AlbumCoverThumbnail: localThumbInput,
			})
		} else {
			err = fmt.Errorf("failed to download for local upload: %v", dlErr)
		}
	}

	if err != nil {
		_, _ = reply.EditText(c, fmt.Sprintf("Failed to send audio: %v", err), nil)
	} else {
		_ = reply.Delete(c, true)
	}

	return gotdbot.EndGroups
}
