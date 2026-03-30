package httpx

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// HasAudioStream checks if a video URL has an audio stream using ffprobe.
func HasAudioStream(url string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-select_streams", "a",
		"-show_entries", "stream=index",
		"-of", "csv=p=0",
		"-user_agent", "Mozilla/5.0",
		url,
	)

	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	return len(strings.TrimSpace(string(output))) > 0, nil
}

// DownloadFileToTemp downloads a file from url and saves it to a temporary file with the given extension.
func DownloadFileToTemp(url, ext string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	tempFile, err := os.CreateTemp("", "snap_*"+ext)
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}

	return tempFile.Name(), nil
}

// DownloadFile downloads a file from url and saves it to filepath.
func DownloadFile(url, targetPath string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	err = os.MkdirAll(filepath.Dir(targetPath), 0755)
	if err != nil {
		return "", err
	}

	out, err := os.Create(targetPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return targetPath, err
}
