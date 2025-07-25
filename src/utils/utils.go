package utils

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"songBot/src/config"
	"strings"
	"time"
)

const (
	defaultDownloadDirPerm = 0755
	defaultFilePerm        = 0644
	maxCoverSize           = 10 << 20 // 10MB
	downloadTimeout        = 4 * time.Minute
)

var (
	errMissingCDNURL         = errors.New("missing CDN URL")
	errMissingKey            = errors.New("missing CDN key")
	errFileNotFound          = errors.New("file not found")
	errInvalidHexKey         = errors.New("invalid hex key")
	errInvalidAESIV          = errors.New("invalid AES IV")
	errVorbisCommentNotFound = errors.New("vorbiscomment not found")
	tgURLRegex               = regexp.MustCompile(`^https:\/\/t\.me\/([a-zA-Z0-9_]{5,})\/(\d+)$`)
)

func (d *Download) processSpotify() (string, []byte, error) {
	track := d.Track
	downloadsDir := config.DownloadPath

	outputFile := filepath.Join(downloadsDir, fmt.Sprintf("%s.ogg", track.TC))
	if _, err := os.Stat(outputFile); err == nil {
		log.Printf("✅ Found existing file: %s", outputFile)
		return outputFile, nil, nil
	}

	if track.Key == "" {
		return "", nil, errMissingKey
	}

	startTime := time.Now()
	defer func() {
		log.Printf("Process completed in %s", time.Since(startTime))
	}()

	// Download and process files
	encryptedFile := filepath.Join(downloadsDir, fmt.Sprintf("%s.encrypted", track.TC))
	decryptedFile := filepath.Join(downloadsDir, fmt.Sprintf("%s_decrypted.ogg", track.TC))

	defer func() {
		_ = os.Remove(encryptedFile)
		_ = os.Remove(decryptedFile)
	}()

	if err := d.downloadAndDecrypt(encryptedFile, decryptedFile); err != nil {
		return "", nil, err
	}

	if err := rebuildOGG(decryptedFile); err != nil {
		log.Printf("Failed to rebuild OGG headers: %v", err)
	}

	return vorbRepairOGG(decryptedFile, track)
}

func (d *Download) downloadAndDecrypt(encryptedPath, decryptedPath string) error {
	// Download file
	resp, err := http.Get(d.Track.CdnURL)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Write encrypted file
	if err := os.WriteFile(encryptedPath, data, defaultFilePerm); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Decrypt file
	decryptedData, decryptTime, err := decryptAudioFile(encryptedPath, d.Track.Key)
	if err != nil {
		return fmt.Errorf("failed to decrypt audio file: %w", err)
	}
	log.Printf("Decryption completed in %s", decryptTime)

	// Write decrypted file
	return os.WriteFile(decryptedPath, decryptedData, defaultFilePerm)
}

func getCover(coverURL string) ([]byte, error) {
	if coverURL == "" {
		return nil, nil
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(coverURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download cover: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Limit reader to prevent OOM with huge files
	coverData, err := io.ReadAll(io.LimitReader(resp.Body, maxCoverSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read cover: %w", err)
	}

	return coverData, nil
}

func decryptAudioFile(filePath, hexKey string) ([]byte, string, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, "", fmt.Errorf("%w: %s", errFileNotFound, filePath)
	}

	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", errInvalidHexKey, err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %w", err)
	}

	audioAesIv, err := hex.DecodeString("72e067fbddcbcf77ebe8bc643f630d93")
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", errInvalidAESIV, err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	startTime := time.Now()
	ctr := cipher.NewCTR(block, audioAesIv)
	decrypted := make([]byte, len(data))
	ctr.XORKeyStream(decrypted, data)

	return decrypted, fmt.Sprintf("%dms", time.Since(startTime).Milliseconds()), nil
}

func rebuildOGG(filename string) error {
	file, err := os.OpenFile(filename, os.O_RDWR, defaultFilePerm)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	writeAt := func(offset int64, data string) error {
		_, err := file.WriteAt([]byte(data), offset)
		return err
	}

	// OGG header structure
	patches := map[int64]string{
		0:  "OggS",
		6:  "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00",
		26: "\x01\x1E\x01vorbis",
		39: "\x02",
		40: "\x44\xAC\x00\x00",
		48: "\x00\xE2\x04\x00",
		56: "\xB8\x01",
		58: "OggS",
		62: "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00",
	}

	for offset, data := range patches {
		if err := writeAt(offset, data); err != nil {
			return fmt.Errorf("failed to write at offset %d: %w", offset, err)
		}
	}

	return nil
}

func vorbRepairOGG(inputFile string, track TrackInfo) (string, []byte, error) {
	coverData, err := getCover(track.Cover)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get cover: %w", err)
	}

	outputFile := filepath.Join(config.DownloadPath, fmt.Sprintf("%s.ogg", track.TC))
	cmd := exec.Command("ffmpeg", "-i", inputFile, "-c", "copy", "-metadata", fmt.Sprintf("lyrics=%s", track.Lyrics), outputFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", coverData, fmt.Errorf("ffmpeg failed: %w\nOutput: %s", err, string(output))
	}

	if err := addVorbisComments(outputFile, track, coverData); err != nil {
		return "", coverData, fmt.Errorf("failed to add vorbis comments: %w", err)
	}

	return outputFile, coverData, nil
}

func addVorbisComments(outputFile string, track TrackInfo, coverData []byte) error {
	if _, err := exec.LookPath("vorbiscomment"); err != nil {
		return errVorbisCommentNotFound
	}

	metadata := fmt.Sprintf(
		"METADATA_BLOCK_PICTURE=%s\n"+
			"ALBUM=%s\n"+
			"ARTIST=%s\n"+
			"TITLE=%s\n"+
			"GENRE=Spotify @FallenProjects\n"+
			"YEAR=%d\n"+
			"TRACKNUMBER=%s\n"+
			"COMMENT=By @FallenProjects\n"+
			"PUBLISHER=%s\n"+
			"DURATION=%d\n",
		createVorbisImageBlock(coverData),
		track.Album,
		track.Artist,
		track.Name,
		track.Year,
		track.TC,
		track.Artist,
		track.Duration,
	)

	tmpFile := "vorbis.txt"
	if err := os.WriteFile(tmpFile, []byte(metadata), defaultFilePerm); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(tmpFile)

	cmd := exec.Command("vorbiscomment", "-a", outputFile, "-c", tmpFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("vorbiscomment failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func createVorbisImageBlock(imageBytes []byte) string {
	tmpCover := "cover.jpg"
	tmpBase64 := "cover.base64"
	defer func() {
		_ = os.Remove(tmpCover)
		_ = os.Remove(tmpBase64)
	}()

	if err := os.WriteFile(tmpCover, imageBytes, defaultFilePerm); err != nil {
		log.Printf("Failed to write cover image: %v", err)
		return ""
	}

	cmd := exec.Command("./cover_gen.sh", tmpCover)
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("Failed to generate cover: %v\nOutput: %s", err, string(output))
		return ""
	}

	data, err := os.ReadFile(tmpBase64)
	if err != nil {
		log.Printf("Failed to read cover data: %v", err)
		return ""
	}

	return string(data)
}

func downloadFile(ctx context.Context, urlStr, filePath string, overwrite bool) (string, error) {
	if urlStr == "" {
		return "", errors.New("empty URL provided")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, downloadTimeout)
	defer cancel()

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Determine filename
	if filePath == "" {
		filePath = determineFilename(urlStr, resp.Header.Get("Content-Disposition"))
	}

	// Skip if file exists and not overwriting
	if !overwrite {
		if _, err := os.Stat(filePath); err == nil {
			return filePath, nil
		}
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), defaultDownloadDirPerm); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Download to temp file first
	tempPath := filePath + ".part"
	if err := writeToFile(tempPath, resp.Body); err != nil {
		return "", err
	}

	// Rename temp file to final name
	if err := os.Rename(tempPath, filePath); err != nil {
		return "", fmt.Errorf("failed to rename temp file: %w", err)
	}

	return filePath, nil
}

func determineFilename(urlStr, contentDisp string) string {
	// Try from Content-Disposition first
	if filename := extractFilename(contentDisp); filename != "" {
		return filepath.Join(config.DownloadPath, sanitizeFilename(filename))
	}

	// Fall back to URL path
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return filepath.Join(config.DownloadPath, uuid.New().String()+".tmp")
	}

	filename := path.Base(parsedURL.Path)
	if filename == "" || filename == "/" || strings.Contains(filename, "?") {
		filename = uuid.New().String() + ".tmp"
	}

	return filepath.Join(config.DownloadPath, sanitizeFilename(filename))
}

func writeToFile(path string, src io.Reader) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	if _, err = io.Copy(file, src); err != nil {
		_ = os.Remove(path)
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func extractFilename(contentDisp string) string {
	re := regexp.MustCompile(`(?i)filename\*?=['"]?(?:UTF-\d['"]*)?([^'";\n]*)['"]?`)
	if match := re.FindStringSubmatch(contentDisp); len(match) > 1 {
		return match[1]
	}
	return ""
}

func sanitizeFilename(name string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			return '_'
		default:
			return r
		}
	}, name)
}

// generateRandomZipName creates a random filename for the ZIP
func generateRandomZipName() string {
	return fmt.Sprintf("playlist_%d.zip", time.Now().Unix())
}
