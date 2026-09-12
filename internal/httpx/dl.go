package httpx

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"
)

// isM3U8 returns true when the URL looks like an HLS stream.
func isM3U8(rawURL string) bool {
	lower := strings.ToLower(rawURL)
	if strings.Contains(lower, ".m3u8") {
		return true
	}
	if u, err := url.Parse(rawURL); err == nil {
		return strings.Contains(strings.ToLower(u.Path), ".m3u8")
	}
	return false
}

func DownloadFile(rawURL string) (string, error) {
	if strings.TrimSpace(rawURL) == "" {
		return "", fmt.Errorf("httpx: empty URL provided")
	}

	template := fmt.Sprintf("dl_%d.%%(ext)s", time.Now().Unix())

	args := []string{
		"--no-warnings",
		"--no-playlist",
		"--restrict-filenames",
		"--no-mtime",
		"--force-overwrites",
		"-o", template,
		"--print", "after_move:filepath",
		"--retries", "1",
		"--merge-output-format", "mp4",
	}

	if isM3U8(rawURL) {
		args = append(args,
			"--hls-prefer-native",
			"--format", "bestvideo+bestaudio/best",
			"--downloader", "ffmpeg",
			"--downloader-args", "ffmpeg:-c copy",
		)
	} else {
		args = append(args, "--format", "bestvideo+bestaudio/best")
	}

	args = append(args, rawURL)

	slog.Info("downloading file", "url", rawURL, "m3u8", isM3U8(rawURL))

	cmd := exec.Command("yt-dlp", args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		slog.Warn("yt-dlp failed", "url", rawURL, "err", err, "output", output)
		return "", fmt.Errorf("httpx: yt-dlp failed: %w\noutput: %s", err, output)
	}

	path, err := extractFilePath(output)
	if err != nil {
		slog.Warn("downloading file failed", "url", rawURL, "err", err)
		return "", fmt.Errorf("httpx: %w\nraw output: %s", err, output)
	}

	slog.Info("download complete", "path", path)
	return path, nil
}

// extractFilePath pulls the last non-empty line from yt-dlp output.
func extractFilePath(output string) (string, error) {
	lines := strings.Split(output, "\n")
	for _, line := range slices.Backward(lines) {
		if line := strings.TrimSpace(line); line != "" {
			return line, nil
		}
	}
	return "", fmt.Errorf("could not extract file path from yt-dlp output")
}

func DownloadImg(rawURL string) (string, error) {
	if strings.TrimSpace(rawURL) == "" {
		return "", fmt.Errorf("httpx: empty URL provided")
	}

	resp, err := client.Get(rawURL)
	if err != nil {
		return "", fmt.Errorf("httpx: failed to GET url: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("httpx: non-200 status code: %d", resp.StatusCode)
	}

	ext := "jpg"
	if strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "png") {
		ext = "png"
	}

	path := fmt.Sprintf("thumb_%d.%s", time.Now().UnixNano(), ext)
	out, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("httpx: failed to create local file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("httpx: failed to write data to file: %w", err)
	}

	return path, nil
}
