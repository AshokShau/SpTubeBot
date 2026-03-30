package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var (
	apiKey string
	apiUrl string
	client = &http.Client{
		Timeout: 30 * time.Second,
	}
	errResp struct {
		Message string `json:"message"`
	}
)

func Init(key, urlStr string) {
	apiKey = key
	apiUrl = urlStr
}

func GetSnap(targetUrl string) (*SnapResponse, error) {
	reqUrl := fmt.Sprintf("%s/api/snap?url=%s", apiUrl, url.QueryEscape(targetUrl))
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Message)
		}
		return nil, fmt.Errorf("API error: status code %d", resp.StatusCode)
	}

	var snapResp SnapResponse
	if err := json.NewDecoder(resp.Body).Decode(&snapResp); err != nil {
		return nil, err
	}

	return &snapResp, nil
}

func Search(targetUrl string) (*SearchResponse, error) {
	reqUrl := fmt.Sprintf("%s/api/search?query=%s", apiUrl, url.QueryEscape(targetUrl))
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Message)
		}
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}
	return &searchResp, nil
}

func GetMusicInfo(targetUrl string) (*SearchResponse, error) {
	reqUrl := fmt.Sprintf("%s/api/get_url?url=%s", apiUrl, url.QueryEscape(targetUrl))
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Message)
		}
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}
	return &searchResp, nil
}

func GetTrack(targetUrl string) (*TrackResponse, error) {
	reqUrl := fmt.Sprintf("%s/api/track?url=%s", apiUrl, url.QueryEscape(targetUrl))
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Message)
		}
	}

	var trackResp TrackResponse
	if err := json.NewDecoder(resp.Body).Decode(&trackResp); err != nil {
		return nil, err
	}
	return &trackResp, nil
}
