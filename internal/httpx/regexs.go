package httpx

import "regexp"

// SnapPatterns contains regex patterns for supported platforms in /api/snap endpoint
var SnapPatterns = map[string]*regexp.Regexp{
	"Instagram": regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?(?:instagram\.com|instagr\.am)\/(?:p|reel|tv|stories\/[A-Za-z0-9_.]+|stories\/highlights)?\/?[A-Za-z0-9._-]*`),
	"TikTok":    regexp.MustCompile(`(?i)https?:\/\/(?:www\.|m\.)?(?:vm\.|vt\.)?tiktok\.com\/[^\s]+`),
	"Pin":       regexp.MustCompile(`(?i)https?:\/\/(?:(?:www\.|[a-z]{2}\.)?pinterest\.[a-z.]+\/pin\/\d+|pin\.it\/[A-Za-z0-9]+)\/?`),
	"X":         regexp.MustCompile(`(?i)https?:\/\/(?:www\.|m\.)?(?:twitter\.com|x\.com)\/[\w._-]+\/status\/\d+`),
	"FaceBook":  regexp.MustCompile(`(?i)https?:\/\/(?:www\.|m\.|web\.)?(?:facebook\.com|fb\.watch|fb\.com)\/.*`),
	"Threads":   regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?threads\.(?:com|net)\/.*`),
	"TwitchClip": regexp.MustCompile(
		`(?i)https?:\/\/(?:www\.|m\.)?(?:` +
			`twitch\.tv\/clip\/[\w-]+|` +
			`clips\.twitch\.tv\/[\w-]+|` +
			`twitch\.tv\/[\w-]+\/clip\/[\w-]+` +
			`)`,
	),
	"KickClip": regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?kick\.com\/[\w._-]+\/clips\/[\w-]+`),
	"SoraAi":   regexp.MustCompile(`^https:\/\/sora\.chatgpt\.com\/p\/s_[0-9a-fA-F]{32}\?psh=[A-Za-z0-9\-_\.]+$`),
	"SunoAi":   regexp.MustCompile(`^https:\/\/suno\.com\/song\/[0-9a-fA-F\-]{36}\/?$`),
	"Reddit":   regexp.MustCompile(`(?i)https?:\/\/(?:www\.|m\.)?reddit\.com\/r\/[\w-]+\/(?:comments\/[\w-]+\/.*|s\/[\w-]+)`),
	"SnapChat": regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?snapchat\.com\/.*`),
}

var YouTubeShortsPattern = regexp.MustCompile(`(?i)https?:\/\/(?:www\.|m\.)?youtube\.com\/shorts\/[a-zA-Z0-9_-]+`)
var YouTubePattern = regexp.MustCompile(`(?i)https?:\/\/(?:www\.|m\.)?(?:youtube\.com\/watch\?v=|youtu\.be\/)[a-zA-Z0-9_-]+`)
var YouTubePostPattern = regexp.MustCompile(`(?i)https?:\/\/(?:www\.|m\.)?youtube\.com\/(?:post\/|channel\/[\w-]+\/community\?lb=)([a-zA-Z0-9_-]+)`)

var MusicPatterns = map[string]*regexp.Regexp{
	"Deezer":      regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?deezer\.com\/(?:[a-z]{2}\/)?(track|album|playlist)\/(\d+)`),
	"SoundCloud":  regexp.MustCompile(`(?i)https?:\/\/(?:(?:www\.|m\.)?soundcloud\.com|on\.soundcloud\.com|snd\.sc)\/.*`),
	"JioSaavn":    regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?jiosaavn\.com\/(song|album|playlist|featured)\/[^\/]+\/([A-Za-z0-9_]+)`),
	"Spotify":     regexp.MustCompile(`(?i)https?:\/\/(?:open\.|www\.)?spotify\.com\/(album|track|playlist|artist)\/([A-Za-z0-9]+)`),
	"Tidal":       regexp.MustCompile(`(?i)https?:\/\/(?:www\.|listen\.)?tidal\.com\/(?:browse\/)?(track|album|playlist)\/([a-zA-Z0-9-]+)`),
	"Gaana":       regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?gaana\.com\/(song|album|playlist|artist)\/([A-Za-z0-9\-]+)`),
	"mxplayer":    regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?mxplayer\.in\/(?:show|movie|shorts)\/.*`),
	"TwitchVideo": regexp.MustCompile(`(?i)https?:\/\/(?:www\.|m\.)?twitch\.tv\/(?:videos|[\w._-]+\/video)\/\d+`),
	"KickVideo":   regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?kick\.com\/[\w._-]+\/videos\/[a-fA-F0-9-]+`),
}
