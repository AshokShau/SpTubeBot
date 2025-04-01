package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"songBot/config"

	"regexp"
	"time"
)

// SpotifyTrackDetails represents details of a Spotify track.
type SpotifyTrackDetails struct {
	CdnURL   string `json:"cdnurl"`
	Key      string `json:"key"`
	Name     string `json:"name"`
	Artist   string `json:"artist"`
	TC       string `json:"tc"`
	Cover    string `json:"cover"`
	Lyrics   string `json:"lyrics"`
	Album    string `json:"album"`
	Year     int    `json:"year"`
	Duration int    `json:"duration"`
}

type Track struct {
	Name       string `json:"name"`
	Artist     string `json:"artist"`
	ID         string `json:"id"`
	Year       string `json:"year"`
	Cover      string `json:"cover"`
	SmallCover string `json:"cover_small"`
	Duration   int    `json:"duration"`
}

type PlatformSong struct {
	Results []Track `json:"results"`
}

type SpotifyData struct {
	SpotifyUrlPattern *regexp.Regexp
	ApiUrl            string
	Client            *http.Client
	Query             string
}

func NewSpotifyData(query string) *SpotifyData {
	return &SpotifyData{
		SpotifyUrlPattern: regexp.MustCompile(`^(https?://)?(open\.spotify\.com/(track|playlist|album|artist)/[a-zA-Z0-9]+)(\?.*)?$`),
		ApiUrl:            config.ApiUrl,
		Client:            &http.Client{Timeout: 10 * time.Second},
		Query:             query,
	}
}

func (s *SpotifyData) IsValid(url string) bool {
	return s.SpotifyUrlPattern.MatchString(url)
}

func (s *SpotifyData) GetInfo(url string) (*PlatformSong, error) {
	if !s.IsValid(url) {
		return nil, errors.New("invalid URL")
	}
	return s.FetchSpotifyData(url)
}

func (s *SpotifyData) FetchSpotifyData(url string) (*PlatformSong, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/get_url_new?url=%s", s.ApiUrl, url), nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return nil, err
	}
	req.Header.Set("X-API-Key", config.ApiKey)

	resp, err := s.Client.Do(req)
	if err != nil {
		log.Printf("HTTP error while fetching Spotify data: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to fetch Spotify data for URL: %s", url)
		return nil, errors.New("failed to fetch data")
	}

	var data PlatformSong
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Printf("JSON decode error: %v", err)
		return nil, err
	}

	return &data, nil
}
func (s *SpotifyData) Search(limit string) (*PlatformSong, error) {
	url := fmt.Sprintf("%s/search_track/%s?lim=%s", s.ApiUrl, s.Query, limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("X-API-Key", config.ApiKey)

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}

	var data PlatformSong
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &data, nil
}
func (s *SpotifyData) GetTrack(trackID string) (*SpotifyTrackDetails, error) {
	url := fmt.Sprintf("%s/get_track?id=%s", s.ApiUrl, trackID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return nil, err
	}
	req.Header.Set("X-API-Key", config.ApiKey)

	resp, err := s.Client.Do(req)
	if err != nil {
		log.Printf("HTTP error while fetching track details: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch track details")
	}

	var data SpotifyTrackDetails
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}
