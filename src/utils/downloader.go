package utils

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Download struct {
	Track TrackInfo
}

// ZipResult contains information about the ZIP creation process
type ZipResult struct {
	ZipPath      string
	SuccessCount int
	Errors       []error
}

// NewDownload creates a new Download instance with proper validation
func NewDownload(track TrackInfo) (*Download, error) {
	if track.CdnURL == "" {
		return nil, errMissingCDNURL
	}
	return &Download{Track: track}, nil
}

// Process handles the download based on the track's platform
func (d *Download) Process() (string, []byte, error) {
	switch {
	case d.Track.CdnURL == "":
		return "", nil, errMissingCDNURL
	case strings.EqualFold(d.Track.Platform, "spotify"):
		return d.processSpotify()
	default:
		return d.processDirectDL()
	}
}

// processDirectDL handles direct downloads with improved error handling
func (d *Download) processDirectDL() (string, []byte, error) {
	track := d.Track

	// Check for Telegram URL pattern
	if tgURLRegex.MatchString(track.CdnURL) {
		coverData, err := getCover(track.Cover)
		if err != nil {
			return track.CdnURL, nil, nil
		}
		return track.CdnURL, coverData, nil
	}

	filePath, err := DownloadFile(context.Background(), track.CdnURL, "", false)
	if err != nil {
		return "", nil, fmt.Errorf("failed to download file: %w", err)
	}

	coverData, err := getCover(track.Cover)
	if err != nil {
		return filePath, nil, fmt.Errorf("failed to get cover: %w", err)
	}

	return filePath, coverData, nil
}

// ZipTracks creates a ZIP archive containing all tracks from PlatformTracks
func ZipTracks(tracks *PlatformTracks) (*ZipResult, error) {
	if len(tracks.Results) == 0 {
		return nil, errors.New("no tracks to process")
	}

	zipFilename := generateRandomZipName()
	zipFile, err := os.Create(zipFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to create zip file: %w", err)
	}

	result := &ZipResult{ZipPath: zipFilename}
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Create a buffered channel to collect file contents
	fileChan := make(chan struct {
		name string
		data []byte
	}, len(tracks.Results))

	// Download all files concurrently
	sem := make(chan struct{}, 10) // Limit concurrent downloads
	errChan := make(chan error, len(tracks.Results))

	for _, track := range tracks.Results {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(t MusicTrack) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			// Download the file and read its contents
			apiData := NewApiData(t.URL)
			trackData, err := apiData.GetTrack()
			if err != nil {
				errChan <- fmt.Errorf("track %s: failed to get track info: %w", t.ID, err)
				return
			}

			dl, err := NewDownload(*trackData)
			if err != nil {
				errChan <- fmt.Errorf("track %s: invalid download: %w", t.ID, err)
				return
			}

			filename, _, err := dl.Process()
			if err != nil {
				errChan <- fmt.Errorf("track %s: failed to download track: %w", t.ID, err)
				return
			}

			defer func() {
				if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
					log.Printf("warning: failed to remove temp file %q: %v", filename, err)
				}
			}()

			data, err := os.ReadFile(filename)
			if err != nil {
				errChan <- fmt.Errorf("track %s: failed to read downloaded file: %w", t.ID, err)
				return
			}

			// Send the data to be written to zip
			fileChan <- struct {
				name string
				data []byte
			}{
				name: filepath.Base(filename),
				data: data,
			}

			mu.Lock()
			result.SuccessCount++
			mu.Unlock()
		}(track)
	}

	wg.Wait()
	close(fileChan)
	close(errChan)

	zipWriter := zip.NewWriter(zipFile)
	defer func() {
		if err := zipWriter.Close(); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to close zip writer: %w", err))
		}
		if err := zipFile.Close(); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to close zip file: %w", err))
		}
	}()

	for file := range fileChan {
		zipEntry, err := zipWriter.Create(file.name)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to create zip entry %q: %w", file.name, err))
			continue
		}

		if _, err := zipEntry.Write(file.data); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to write %q to zip: %w", file.name, err))
			continue
		}
	}

	for err := range errChan {
		result.Errors = append(result.Errors, err)
	}

	if absPath, err := filepath.Abs(zipFilename); err == nil {
		result.ZipPath = absPath
	}

	if result.SuccessCount == 0 {
		return result, fmt.Errorf("no tracks were successfully added to the zip: %v", result.Errors)
	}

	return result, nil
}
