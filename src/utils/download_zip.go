package utils

import (
	"archive/zip"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// ZipResult contains information about the ZIP creation process
type ZipResult struct {
	ZipPath      string
	SuccessCount int
	Errors       []error
}

// ZipTracks creates a ZIP archive containing all tracks from PlatformTracks
func ZipTracks(tracks *PlatformTracks) (*ZipResult, error) {
	zipFilename := generateRandomZipName()
	result := &ZipResult{ZipPath: zipFilename}
	zipFile, err := os.Create(zipFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to create zip file: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, track := range tracks.Results {
		err := processTrack(zipWriter, track)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("track %s: %v", track.ID, err))
			continue
		}
		result.SuccessCount++
	}

	if absPath, err := filepath.Abs(zipFilename); err == nil {
		result.ZipPath = absPath
	}

	if result.SuccessCount == 0 {
		return result, fmt.Errorf("no tracks were successfully added to the zip")
	}

	return result, nil
}

// processTrack handles downloading and adding a single track to the ZIP
func processTrack(zipWriter *zip.Writer, track MusicTrack) error {
	// Get the track data
	apiData := NewApiData(track.URL)
	trackData, err := apiData.GetTrack()
	if err != nil {
		return fmt.Errorf("failed to get track info: %v", err)
	}

	filename, _, err := NewDownload(*trackData).Process()
	if err != nil {
		return fmt.Errorf("failed to download track: %v", err)
	}

	audioFile, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open downloaded file: %v", err)
	}
	defer audioFile.Close()

	baseName := filepath.Base(filename)
	zipEntry, err := zipWriter.Create(baseName)
	if err != nil {
		return fmt.Errorf("failed to create zip entry: %v", err)
	}

	if _, err := io.Copy(zipEntry, audioFile); err != nil {
		return fmt.Errorf("failed to write to zip: %v", err)
	}

	defer func() {
		_ = os.Remove(filename)
	}()

	return nil
}

// generateRandomZipName creates a random filename for the ZIP
func generateRandomZipName() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("tracks_%x_%d.zip", b, time.Now().Unix())
}
