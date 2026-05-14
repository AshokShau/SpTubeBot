package commands

import (
	"bytes"
	"fmt"
	"html"
	"noinoi/internal/database"
	"noinoi/internal/httpx"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AshokShau/gotdbot"
)

func getYouTubeUrl(m *gotdbot.Message) (string, string) {
	text := m.GetText()
	if text == "" {
		return "", ""
	}

	if match := httpx.YouTubeShortsPattern.FindString(text); match != "" {
		return match, "short"
	}

	if match := httpx.YouTubePostPattern.FindString(text); match != "" {
		return match, "post"
	}

	if match := httpx.YouTubePattern.FindString(text); match != "" {
		return match, "video"
	}

	return "", ""
}

func downloadYouTube(url string, audioOnly bool) (string, string, string, string, error) {
	tempDir, err := os.MkdirTemp("", "ytdl_*")
	if err != nil {
		return "", "", "", "", err
	}

	outputTemplate := filepath.Join(tempDir, "%(title).200s.%(ext)s")
	thumbTemplate := filepath.Join(tempDir, "thumb.%(ext)s")

	args := []string{
		"--no-playlist",
		"--match-filter", "duration <= 7200",
		"--print", "%(title)s",
		"--print", "after_move:%(filepath)s",
		"--write-thumbnail",
		"--convert-thumbnails", "jpg",
		"-o", outputTemplate,
		"-o", "thumbnail:" + thumbTemplate,
	}

	if audioOnly {
		args = append(args, "-f", "bestaudio[ext=m4a]/bestaudio", "--extract-audio", "--audio-format", "m4a")
	} else {
		args = append(args, "-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best")
	}
	args = append(args, url)

	cmd := exec.Command("yt-dlp", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	stdoutStr := stdout.String()
	stderrStr := stderr.String()

	if strings.Contains(stderrStr, "does not pass filter") || strings.Contains(stdoutStr, "does not pass filter") {
		os.RemoveAll(tempDir)
		return "", "", "", "", fmt.Errorf("DURATION_EXCEEDED")
	}

	if err != nil {
		os.RemoveAll(tempDir)
		return "", "", "", "", fmt.Errorf("failed to download: %v (stderr: %s)", err, stderrStr)
	}

	lines := strings.Split(strings.TrimSpace(stdoutStr), "\n")
	if len(lines) < 2 {
		os.RemoveAll(tempDir)
		return "", "", "", "", fmt.Errorf("failed to extract title or path from output: %s", stdoutStr)
	}

	title := lines[0]
	actualPath := lines[1]

	thumbPath := filepath.Join(tempDir, "thumb.jpg")
	if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
		thumbPath = ""
	}

	return actualPath, thumbPath, title, tempDir, nil
}

func youtubeHandler(c *gotdbot.Client, ctx *gotdbot.Context) error {
	m := ctx.EffectiveMessage
	if m.IsCommand() {
		return nil
	}

	url, typ := getYouTubeUrl(m)
	if url == "" {
		return nil
	}

	c.Logger.Info("YouTube download request (auto)", "user_id", m.SenderID(), "chat_id", m.ChatId, "url", url)

	botId := c.Me.Id

	reply, err := m.ReplyText(c, "⏳ Processing YouTube...", nil)
	if err != nil {
		return err
	}

	if typ == "post" {
		data, err := httpx.GetYouTubePost(url)
		if err != nil {
			_, _ = reply.EditText(c, fmt.Sprintf("Error: %v", err), nil)
			return nil
		}

		if len(data.Images) == 0 {
			_, _ = reply.EditText(c, "No images found in this YouTube post.", nil)
			return nil
		}

		caption := "Join @FallenProjects"
		if data.Text != "" {
			text := data.Text
			if len(text) > 700 {
				text = text[:700] + "..."
			}
			caption = fmt.Sprintf("<b>%s</b>\n\nJoin @FallenProjects", html.EscapeString(text))
		}

		if len(data.Images) == 1 {
			_, err = handleMediaUpload(c, m, SnapMediaItem{URL: data.Images[0]}, "photo", caption)
		} else {
			images := data.Images
			if len(images) > 10 {
				images = images[:10]
			}
			var items []SnapMediaItem
			for _, img := range images {
				items = append(items, SnapMediaItem{URL: img})
			}
			err = sendMediaAlbum(c, m, items, "photo", caption)
		}

		if err != nil {
			_, _ = reply.EditText(c, fmt.Sprintf("Failed to upload: %v", err), nil)
		} else {
			database.IncrementDownloads(botId, true)
			_ = reply.Delete(c, true)
		}
		return gotdbot.EndGroups
	}

	audioOnly := typ == "video"
	filePath, thumbPath, title, tempDir, err := downloadYouTube(url, audioOnly)
	if err != nil {
		if err.Error() == "DURATION_EXCEEDED" {
			_, _ = reply.EditText(c, "Sorry, videos over 1 hour are not supported.", nil)
			return gotdbot.EndGroups
		}

		_, _ = reply.EditText(c, fmt.Sprintf("Error: %v", err), nil)
		return nil
	}

	defer os.RemoveAll(tempDir)

	escapedTitle := html.EscapeString(title)
	caption := fmt.Sprintf("<b>%s</b>\n\nJoin @FallenProjects", escapedTitle)
	input := &gotdbot.InputFileLocal{Path: filePath}

	var thumbInput *gotdbot.InputThumbnail
	if thumbPath != "" {
		thumbInput = &gotdbot.InputThumbnail{Thumbnail: &gotdbot.InputFileLocal{Path: thumbPath}}
	}

	if audioOnly {
		_, err = m.ReplyAudio(c, input, &gotdbot.SendAudioOpts{
			Caption:             caption,
			ParseMode:           "HTML",
			Title:               title,
			AlbumCoverThumbnail: thumbInput,
		})
	} else {
		_, err = m.ReplyVideo(c, input, &gotdbot.SendVideoOpts{
			Caption:   caption,
			ParseMode: "HTML",
			Thumbnail: thumbInput,
		})
	}

	if err != nil {
		_, _ = reply.EditText(c, fmt.Sprintf("Failed to upload: %v", err), nil)
	} else {
		database.IncrementDownloads(botId, true)
		_ = reply.Delete(c, true)
	}

	return gotdbot.EndGroups
}

func ytCommandHandler(c *gotdbot.Client, ctx *gotdbot.Context) error {
	m := ctx.EffectiveMessage
	url := getUrl(m)
	if url == "" {
		_, _ = m.ReplyText(c, "Usage: /yt <url>", nil)
		return nil
	}

	c.Logger.Info("YouTube download request (command)", "user_id", m.SenderID(), "chat_id", m.ChatId, "url", url)

	botId := c.Me.Id

	reply, err := m.ReplyText(c, "⏳ Processing YouTube Video...", nil)
	if err != nil {
		return err
	}

	filePath, thumbPath, title, tempDir, err := downloadYouTube(url, false)
	if err != nil {
		if err.Error() == "DURATION_EXCEEDED" {
			_, _ = reply.EditText(c, "Sorry, videos over 1 hour are not supported.", nil)
			return gotdbot.EndGroups
		}

		_, _ = reply.EditText(c, fmt.Sprintf("Error: %v", err), nil)
		return nil
	}
	defer os.RemoveAll(tempDir)

	escapedTitle := html.EscapeString(title)
	caption := fmt.Sprintf("<b>%s</b>\n\nJoin @FallenProjects", escapedTitle)
	input := &gotdbot.InputFileLocal{Path: filePath}

	var thumbInput *gotdbot.InputThumbnail
	if thumbPath != "" {
		thumbInput = &gotdbot.InputThumbnail{Thumbnail: &gotdbot.InputFileLocal{Path: thumbPath}}
	}

	_, err = m.ReplyVideo(c, input, &gotdbot.SendVideoOpts{
		Caption:   caption,
		ParseMode: "HTML",
		Thumbnail: thumbInput,
	})

	if err != nil {
		_, _ = reply.EditText(c, fmt.Sprintf("Failed to upload: %v", err), nil)
	} else {
		database.IncrementDownloads(botId, true)
		_ = reply.Delete(c, true)
	}

	return nil
}
