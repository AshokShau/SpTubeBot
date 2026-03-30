package httpx

type Video struct {
	URL       string `json:"url,omitempty"`
	Thumbnail string `json:"thumbnail,omitempty"`
}

type Audio struct {
	URL string `json:"url,omitempty"`
}

// SnapResponse for /api/snap endpoint
type SnapResponse struct {
	Videos []Video  `json:"videos,omitempty"`
	Audios []Audio  `json:"audios,omitempty"`
	Images []string `json:"images,omitempty"`
	Title  string   `json:"title"`
}

type Track struct {
	Title     string `json:"title"`
	ID        string `json:"id"`
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
	Duration  int    `json:"duration"`
	Channel   string `json:"channel"`
	Views     string `json:"views"`
	Platform  string `json:"platform"`
}

// SearchResponse for /api/get_url or /api/search endpoint
type SearchResponse struct {
	Results []Track `json:"results"`
}

// TrackResponse for /api/track endpoint
type TrackResponse struct {
	Id       string `json:"id"`
	URL      string `json:"url"`
	CdnURL   string `json:"cdnurl"`
	Key      string `json:"key,omitempty"`
	Platform string `json:"platform"`
}
