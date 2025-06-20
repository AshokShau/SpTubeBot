package utils

// TrackInfo represents details of a Spotify track.
type TrackInfo struct {
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
	Platform string `json:"platform"`
}

type MusicTrack struct {
	Name       string `json:"name"`
	Artist     string `json:"artist"`
	ID         string `json:"id"`
	URL        string `json:"url"`
	Year       string `json:"year"`
	Cover      string `json:"cover"`
	SmallCover string `json:"cover_small"`
	Duration   int    `json:"duration"`
	Platform   string `json:"platform"`
}

type PlatformTracks struct {
	Results []MusicTrack `json:"results"`
}
