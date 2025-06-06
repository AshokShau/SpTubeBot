package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Download struct {
	Track SpotifyTrackDetails
}

func NewDownload(track SpotifyTrackDetails) *Download {
	return &Download{Track: track}
}

const (
	oggS        = "OggS"
	zeroes      = "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"
	vorbisStart = "\x01\x1E\x01vorbis"
	channels    = "\x02"
	sampleRate  = "\x44\xAC\x00\x00"
	bitRate     = "\x00\xE2\x04\x00"
	packetSizes = "\xB8\x01"
)

func rebuildOGG(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Printf("❌ Error: %s not found.\n", filename)
		return err
	}

	file, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		log.Printf("❌ Error opening file: %v\n", err)
		return err
	}
	defer file.Close()

	writeAt := func(offset int64, data string) error {
		_, err := file.WriteAt([]byte(data), offset)
		return err
	}

	if err := writeAt(0, oggS); err != nil {
		return err
	}
	if err := writeAt(6, zeroes); err != nil {
		return err
	}
	if err := writeAt(26, vorbisStart); err != nil {
		return err
	}
	if err := writeAt(39, channels); err != nil {
		return err
	}
	if err := writeAt(40, sampleRate); err != nil {
		return err
	}
	if err := writeAt(48, bitRate); err != nil {
		return err
	}
	if err := writeAt(56, packetSizes); err != nil {
		return err
	}
	if err := writeAt(58, oggS); err != nil {
		return err
	}
	if err := writeAt(62, zeroes); err != nil {
		return err
	}

	return nil
}

func decryptAudioFile(filePath string, hexKey string) ([]byte, string, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, "", fmt.Errorf("file not found: %v", err)
	}

	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, "", fmt.Errorf("invalid hex key: %v", err)
	}

	buffer, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %v", err)
	}

	audioAesIv, err := hex.DecodeString("72e067fbddcbcf77ebe8bc643f630d93")
	if err != nil {
		return nil, "", fmt.Errorf("invalid AES IV: %v", err)
	}
	ivInt := int64(0)
	for i, b := range audioAesIv {
		ivInt |= int64(b) << (8 * (len(audioAesIv) - i - 1))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create AES cipher: %v", err)
	}

	ctr := cipher.NewCTR(block, audioAesIv)
	startTime := time.Now()
	decryptedBuffer := make([]byte, len(buffer))
	ctr.XORKeyStream(decryptedBuffer, buffer)

	decryptTime := time.Since(startTime).Milliseconds()
	return decryptedBuffer, fmt.Sprintf("%dms", decryptTime), nil
}

func (d *Download) Process() (string, []byte, error) {
	downloadsDir := "downloads"
	track := d.Track
	encryptedFile := filepath.Join(downloadsDir, fmt.Sprintf("%s.encrypted", track.TC))
	dcryptedFile := filepath.Join(downloadsDir, fmt.Sprintf("%s_decrypted.ogg", track.TC))
	outputFile := filepath.Join(downloadsDir, fmt.Sprintf("%s.ogg", track.TC))

	if _, err := os.Stat(outputFile); err == nil {
		log.Printf("✅ Found existing file: %s", outputFile)
		return outputFile, nil, nil
	}

	if track.CdnURL == "" || track.Key == "" {
		return "", nil, fmt.Errorf("missing CDN URL or key")
	}

	startTime := time.Now()
	rawFile, err := http.Get(track.CdnURL)
	if err != nil {
		return "", nil, fmt.Errorf("failed to download file: %v", err)
	}
	defer rawFile.Body.Close()
	buffer, _ := io.ReadAll(rawFile.Body)

	err = os.WriteFile(encryptedFile, buffer, 0644)
	if err != nil {
		return "", nil, fmt.Errorf("failed to write file: %v", err)
	}

	defer os.Remove(encryptedFile)

	decryptedBuffer, decryptTime, err := decryptAudioFile(encryptedFile, track.Key)
	if err != nil {
		return "", nil, fmt.Errorf("failed to decrypt audio file: %v", err)
	}

	log.Printf("Decryption completed in %s", decryptTime)

	err = os.WriteFile(dcryptedFile, decryptedBuffer, 0644)
	if err != nil {
		return "", nil, fmt.Errorf("failed to write decrypted file: %v", err)
	}

	defer os.Remove(dcryptedFile)

	err = rebuildOGG(dcryptedFile)
	if err != nil {
		log.Printf("Failed to rebuild OGG headers: %v", err)
	}

	_, err = exec.LookPath("vorbiscomment")
	if err != nil {
		return "", nil, fmt.Errorf("vorbiscomment not found: %v", err)
	}

	fixedFile, thumb, err := vorbRepairOGG(dcryptedFile, track)
	if err != nil {
		return "", nil, fmt.Errorf("failed to repair audio file: %v", err)
	}

	// defer os.Remove(fixedFile)

	log.Printf("Process completed in %s", time.Since(startTime))
	return fixedFile, thumb, nil
}

func vorbRepairOGG(inputFile string, r SpotifyTrackDetails) (string, []byte, error) {
	cov, err := http.Get(r.Cover)
	if err != nil {
		return "", nil, fmt.Errorf("failed to download cover: %w", err)
	}
	defer cov.Body.Close()

	coverData, err := io.ReadAll(cov.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read cover: %w", err)
	}

	outputFile := filepath.Join("downloads", fmt.Sprintf("%s.ogg", r.TC))
	cmd := exec.Command("ffmpeg", "-i", inputFile, "-c", "copy", "-metadata", fmt.Sprintf("lyrics=%s", r.Lyrics), outputFile)

	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", nil, fmt.Errorf("ffmpeg failed: %w\nOutput: %s", err, string(output))
	}

	vorbisMetadata := fmt.Sprintf(
		"METADATA_BLOCK_PICTURE=%s\n"+"ALBUM=%s\n"+"ARTIST=%s\n"+"TITLE=%s\n"+"GENRE=Spotify @FallenProjects @BeatsXBot\n"+"YEAR=%d\n"+"TRACKNUMBER=%s\n"+"COMMENT=Downloaded via @BeatsXBot by @FallenProjects\n"+"PUBLISHER=%s\n"+"DURATION=%d\n",
		createVorbisImageBlock(coverData), r.Album, r.Artist, r.Name, r.Year, r.TC, r.Artist, r.Duration,
	)

	tmpFile := "vorbis.txt"
	if err := os.WriteFile(tmpFile, []byte(vorbisMetadata), 0644); err != nil {
		return "", coverData, fmt.Errorf("failed to write Vorbis metadata: %w", err)
	}
	defer os.Remove(tmpFile)

	cmd = exec.Command("vorbiscomment", "-a", outputFile, "-c", tmpFile)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return "", coverData, fmt.Errorf("vorbiscomment failed: %w\nOutput: %s", err, string(output))
	}

	return outputFile, coverData, nil
}

func createVorbisImageBlock(imageBytes []byte) string {
	os.WriteFile("cover.jpg", imageBytes, 0644)
	defer os.Remove("cover.jpg")
	cmd := exec.Command("./cover_gen.sh", "cover.jpg")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to execute cover_gen.sh: %v\nOutput: %s", err, string(output))
		return ""
	}
	coverData, err := os.ReadFile("cover.base64")
	if err != nil {
		log.Printf("Failed to read cover data: %v", err)
		return ""
	}
	defer os.Remove("cover.base64")
	return string(coverData)
}
