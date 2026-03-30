package commands

import (
	"bytes"
	"fmt"
	"html"
	"noinoi/internal/httpx"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AshokShau/gotdbot"
)

func getYouTubeUrl(m *gotdbot.Message) (string, bool) {
	text := m.GetText()
	if text == "" {
		return "", false
	}

	if match := httpx.YouTubeShortsPattern.FindString(text); match != "" {
		return match, true
	}

	if match := httpx.YouTubePattern.FindString(text); match != "" {
		return match, false
	}

	return "", false
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
		"--match-filter", "duration <= 3600",
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

	url, isShort := getYouTubeUrl(m)
	if url == "" {
		return nil
	}

	reply, err := m.ReplyText(c, "⏳ Processing YouTube...", nil)
	if err != nil {
		return err
	}

	audioOnly := !isShort
	filePath, thumbPath, title, tempDir, err := downloadYouTube(url, audioOnly)
	if err != nil {
		if err.Error() == "DURATION_EXCEEDED" {
			_, _ = reply.EditText(c, "Sorry, videos over 1 hour are not supported.", nil)
		} else {
			_, _ = reply.EditText(c, fmt.Sprintf("Error: %v", err), nil)
		}
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

	reply, err := m.ReplyText(c, "⏳ Processing YouTube Video...", nil)
	if err != nil {
		return err
	}

	filePath, thumbPath, title, tempDir, err := downloadYouTube(url, false)
	if err != nil {
		if err.Error() == "DURATION_EXCEEDED" {
			_, _ = reply.EditText(c, "Sorry, videos over 1 hour are not supported.", nil)
		} else {
			_, _ = reply.EditText(c, fmt.Sprintf("Error: %v", err), nil)
		}
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
		_ = reply.Delete(c, true)
	}

	return nil
}
