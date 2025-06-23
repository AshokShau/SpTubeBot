package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"songBot/src/config"
)

// Constants for API configuration
const (
	apiTimeout     = 60 * time.Second
	defaultLimit   = "10"
	maxQueryLength = 100
	maxURLLength   = 2048
)

// URL patterns for supported platforms
var urlPatterns = map[string]*regexp.Regexp{
	"spotify":       regexp.MustCompile(`^(https?://)?(open\.spotify\.com/(track|playlist|album|artist)/[a-zA-Z0-9]+)(\?.*)?$`),
	"youtube":       regexp.MustCompile(`^(https?://)?(www\.)?(youtube\.com/watch\?v=|youtu\.be/)[a-zA-Z0-9_-]+(\?.*)?$`),
	"youtube_music": regexp.MustCompile(`^(https?://)?(music\.)?youtube\.com/(watch\?v=|playlist\?list=)[a-zA-Z0-9_-]+(\?.*)?$`),
	"soundcloud":    regexp.MustCompile(`^(https?://)?(www\.)?soundcloud\.com/[a-zA-Z0-9_-]+(/(sets)?/[a-zA-Z0-9_-]+)?(\?.*)?$`),
	"apple_music":   regexp.MustCompile(`^(https?://)?(music|geo)\.apple\.com/[a-z]{2}/(album|playlist|song)/[^/]+/[0-9]+(\?i=[0-9]+)?(\?.*)?$`),
}

// ApiData represents the API client and configuration
type ApiData struct {
	ApiUrl string
	Client *http.Client
	Query  string
}

// NewApiData creates a new ApiData instance with proper initialization
func NewApiData(query string) *ApiData {
	return &ApiData{
		ApiUrl: config.ApiUrl,
		Client: &http.Client{Timeout: apiTimeout},
		Query:  sanitizeInput(query),
	}
}

// IsValid checks if the URL is valid and supported
func (s *ApiData) IsValid(rawURL string) bool {
	if rawURL == "" || len(rawURL) > maxURLLength {
		return false
	}

	// Basic URL format validation
	if _, err := url.ParseRequestURI(rawURL); err != nil {
		return false
	}

	for _, pattern := range urlPatterns {
		if pattern.MatchString(rawURL) {
			return true
		}
	}
	return false
}

// GetInfo fetches track information for a given URL
func (s *ApiData) GetInfo(rawURL string) (*PlatformTracks, error) {
	if !s.IsValid(rawURL) {
		return nil, errors.New("invalid or unsupported URL")
	}

	return s.FetchData(rawURL)
}

// FetchData makes the actual API request to fetch data
func (s *ApiData) FetchData(rawURL string) (*PlatformTracks, error) {
	apiURL := fmt.Sprintf("%s/get_url?url=%s", s.ApiUrl, url.QueryEscape(rawURL))
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-API-Key", config.ApiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var data PlatformTracks
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &data, nil
}

// Search performs a track search with the given query
func (s *ApiData) Search(limit string) (*PlatformTracks, error) {
	if limit == "" {
		limit = defaultLimit
	}

	searchURL := fmt.Sprintf("%s/search_track/%s?lim=%s",
		s.ApiUrl,
		url.QueryEscape(s.Query),
		url.QueryEscape(limit))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-API-Key", config.ApiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var data PlatformTracks
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &data, nil
}

// GetTrack fetches detailed information for a specific track
func (s *ApiData) GetTrack(trackID string) (*TrackInfo, error) {
	if trackID == "" {
		return nil, errors.New("empty track ID")
	}

	trackURL := fmt.Sprintf("%s/get_track?id=%s", s.ApiUrl, url.QueryEscape(trackID))

	req, err := http.NewRequest("GET", trackURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-API-Key", config.ApiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var data TrackInfo
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &data, nil
}

// sanitizeInput performs basic input sanitization
func sanitizeInput(input string) string {
	if len(input) > maxQueryLength {
		return input[:maxQueryLength]
	}
	return input
}
