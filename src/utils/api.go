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

// Constants for API configuration and validation
const (
	apiTimeout      = 60 * time.Second
	defaultLimit    = "10"
	maxQueryLength  = 500
	maxURLLength    = 5000
	headerAccept    = "Accept"
	headerAPIKey    = "X-API-Key"
	mimeApplication = "application/json"
)

// URL patterns to detect supported music platforms
var urlPatterns = map[string]*regexp.Regexp{
	"spotify":       regexp.MustCompile(`^(https?://)?(open\.spotify\.com/(track|playlist|album|artist)/[a-zA-Z0-9]+)(\?.*)?$`),
	"youtube":       regexp.MustCompile(`^(https?://)?(www\.)?(youtube\.com/watch\?v=|youtu\.be/)[\w-]+(\?.*)?$`),
	"youtube_music": regexp.MustCompile(`^(https?://)?(music\.)?youtube\.com/(watch\?v=|playlist\?list=)[\w-]+(\?.*)?$`),
	"soundcloud":    regexp.MustCompile(`^(https?://)?(www\.)?soundcloud\.com/[\w-]+(/[\w-]+)?(/sets/[\w-]+)?(\?.*)?$`),
	"apple_music":   regexp.MustCompile(`^(https?://)?(music|geo)\.apple\.com/[a-z]{2}/(album|playlist|song)/[^/]+/\d+(\?i=\d+)?(\?.*)?$`)}

// ApiData represents a reusable HTTP client for API operations
type ApiData struct {
	ApiUrl string
	Client *http.Client
	Query  string
}

// NewApiData creates and returns an ApiData instance
func NewApiData(query string) *ApiData {
	return &ApiData{
		ApiUrl: config.ApiUrl,
		Client: &http.Client{Timeout: apiTimeout},
		Query:  sanitizeInput(query),
	}
}

// IsValid checks if the provided URL is valid and belongs to a supported platform
func (api *ApiData) IsValid(rawURL string) bool {
	if rawURL == "" || len(rawURL) > maxURLLength {
		return false
	}

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

// GetInfo fetches track or playlist details from a given URL
func (api *ApiData) GetInfo() (*PlatformTracks, error) {
	rawURL := api.Query
	if !api.IsValid(rawURL) {
		return nil, errors.New("invalid or unsupported URL")
	}
	return api.FetchData(rawURL)
}

// FetchData performs a GET request to /get_url to retrieve platform metadata
func (api *ApiData) FetchData(rawURL string) (*PlatformTracks, error) {
	endpoint := fmt.Sprintf("%s/get_url?url=%s", api.ApiUrl, url.QueryEscape(rawURL))
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %w", err)
	}

	api.setHeaders(req)

	resp, err := api.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result PlatformTracks
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("JSON decode failed: %w", err)
	}

	return &result, nil
}

// Search performs a keyword-based song search on the API
func (api *ApiData) Search(limit string) (*PlatformTracks, error) {
	if limit == "" {
		limit = defaultLimit
	}

	endpoint := fmt.Sprintf("%s/search_track/%s?lim=%s",
		api.ApiUrl,
		url.QueryEscape(api.Query),
		url.QueryEscape(limit),
	)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %w", err)
	}

	api.setHeaders(req)

	resp, err := api.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result PlatformTracks
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("JSON decode failed: %w", err)
	}

	return &result, nil
}

// GetTrack fetches metadata for a specific track by its ID
func (api *ApiData) GetTrack() (*TrackInfo, error) {
	trackID := api.Query
	if trackID == "" {
		return nil, errors.New("empty track ID")
	}

	endpoint := fmt.Sprintf("%s/get_track?id=%s", api.ApiUrl, url.QueryEscape(trackID))
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %w", err)
	}

	api.setHeaders(req)

	resp, err := api.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var track TrackInfo
	if err := json.NewDecoder(resp.Body).Decode(&track); err != nil {
		return nil, fmt.Errorf("JSON decode failed: %w", err)
	}

	return &track, nil
}

// setHeaders sets common headers on the HTTP request
func (api *ApiData) setHeaders(req *http.Request) {
	req.Header.Set(headerAPIKey, config.ApiKey)
	req.Header.Set(headerAccept, mimeApplication)
}

// sanitizeInput trims overly long queries
func sanitizeInput(input string) string {
	if len(input) > maxQueryLength {
		return input[:maxQueryLength]
	}
	return input
}
