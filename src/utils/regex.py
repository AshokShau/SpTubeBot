import re

# === URL Regex Patterns ===
URL_PATTERNS = {
    "spotify": re.compile(
        r'^(https?://)?([a-z0-9-]+\.)*spotify\.com/(track|playlist|album|artist)/[a-zA-Z0-9]+(\?.*)?$'),
    "youtube": re.compile(r'^(https?://)?([a-z0-9-]+\.)*(youtube\.com/watch\?v=|youtu\.be/)[\w-]+(\?.*)?$'),
    "youtube_music": re.compile(r'^(https?://)?([a-z0-9-]+\.)*youtube\.com/(watch\?v=|playlist\?list=)[\w-]+(\?.*)?$'),
    "soundcloud": re.compile(r'^(https?://)?([a-z0-9-]+\.)*soundcloud\.com/[\w-]+(/[\w-]+)?(/sets/[\w-]+)?(\?.*)?$'),
    "apple_music": re.compile(
        r'^(https?://)?([a-z0-9-]+\.)?apple\.com/[a-z]{2}/(album|playlist|song)/[^/]+/(pl\.[a-zA-Z0-9]+|\d+)(\?i=\d+)?(\?.*)?$')
}

# === Save Snap Regex Patterns ===
SAVE_SNAP_PATTERNS = [

    # Instagram
    re.compile(
        r"(?i)https?://(?:www\.)?(?:instagram\.com|instagr\.am)/(?:p|reel|reels|tv|stories/(?:highlights/)?[\w._-]+)?/?[\w._-]*"
    ),

    # TikTok
    re.compile(
        r"(?i)https?://(?:www\.|m\.)?(?:"
        r"tiktok\.com/@[\w.-]+/(?:video|photo)/\d+|"
        r"tiktok\.com/(?:t|v)/\d+|"
        r"vm\.tiktok\.com/[\w-]+|"
        r"vt\.tiktok\.com/[\w-]+)"
    ),

    # Pinterest
    re.compile(
        r"(?i)https?://(?:www\.|[a-z]{2}\.)?(?:pinterest\.[a-z.]+/pin/\d+|pin\.it/[\w-]+)/?"
    ),

    # Twitter / X
    re.compile(
        r"(?i)https?://(?:www\.|m\.)?(?:twitter\.com|x\.com)/[\w._-]+/status/\d+"
    ),

    # Snapchat
    re.compile(
        r"(?i)https?://(?:www\.)?snapchat\.com/.*"
    ),

    # Facebook (fb.watch + videos/reels)
    re.compile(
        r"(?i)https?://(?:www\.|m\.|web\.)?(?:facebook\.com|fb\.watch|fb\.com)/.*"
    ),

    # LinkedIn
    re.compile(
        r"(?i)https?://(?:www\.)?linkedin\.com/.*"
    ),

    # SoundCloud
    re.compile(
        r"(?i)https?://(?:www\.|m\.)?soundcloud\.com/.*"
    ),

    # Bilibili
    re.compile(
        r"(?i)https?://(?:www\.|m\.)?bilibili\.com/video/.*"
    ),

    # CapCut
    re.compile(
        r"(?i)https?://(?:www\.|m\.)?capcut\.com/.*"
    ),

    # IMDb video
    re.compile(
        r"(?i)https?://(?:www\.|m\.)?imdb\.com/video/.*"
    ),

    # ShareChat
    re.compile(
        r"(?i)https?://(?:www\.)?sharechat\.com/.*"
    ),

    # Streamable
    re.compile(
        r"(?i)https?://(?:www\.)?streamable\.com/.*"
    ),

    # TED Talks
    re.compile(
        r"(?i)https?://(?:www\.)?ted\.com/talks/.*"
    ),

    # Threads
    re.compile(
        r"(?i)https?://(?:www\.)?threads\.(?:net|com)/.*"
    ),

    # VK
    re.compile(
        r"(?i)https?://(?:www\.|m\.)?vk\.com/.*"
    ),

    # Twitch VOD
    re.compile(
        r"(?i)https?://(?:www\.|m\.)?twitch\.tv/(?:videos|[\w._-]+/video)/\d+"
    ),

    # Twitch Clips
    re.compile(
        r"(?i)https?://(?:www\.|m\.)?(?:"
        r"twitch\.tv/clip/[\w-]+|"
        r"clips\.twitch\.tv/[\w-]+)"
    ),

    # Tumblr
    re.compile(
        r"(?i)https?://(?:www\.|m\.)?tumblr\.com/.*"
    ),

    # Dailymotion
    re.compile(
        r"(?i)https?://(?:www\.)?dailymotion\.com/video/.*"
    ),

    # Douyin
    re.compile(
        r"(?i)https?://(?:www\.|v\.)?douyin\.com/.*"
    ),

    # 9GAG
    re.compile(
        r"(?i)https?://(?:www\.)?9gag\.com/gag/[\w-]+"
    ),

    # AkilliTV
    re.compile(
        r"(?i)https?://(?:www\.)?akillitv\.com/(?:video|embed)/[\w-]+"
    ),

    # Bandcamp
    re.compile(
        r"(?i)https?://(?:www\.|m\.)?[\w.-]+\.bandcamp\.com/(?:track|album|video)/[\w-]+"
    ),

    # BitChute
    re.compile(
        r"(?i)https?://(?:www\.)?bitchute\.com/video/[\w-]+"
    ),

    # Blogger
    re.compile(
        r"(?i)https?://(?:www\.)?blogger\.com/.*"
    ),

    # Rumble
    re.compile(
        r"(?i)https?://(?:www\.)?rumble\.com/.*"
    ),

    # ESPN
    re.compile(
        r"(?i)https?://(?:www\.)?espn\.com/video/.*"
    ),

    # BuzzFeed
    re.compile(
        r"(?i)https?://(?:www\.)?buzzfeed\.com/.*"
    ),

    # Mastodon
    re.compile(
        r"(?i)https?://(?:[\w.-]+\.)?mastodon\.social/@[\w.-]+/\d+"
    ),

    # Imgur
    re.compile(
        r"(?i)https?://(?:www\.)?imgur\.com/(?:gallery|a|t)?/?[\w-]+"
    ),

    # Reddit (full + short)
    re.compile(
        r"(?i)https?://(?:www\.|m\.)?reddit\.com/r/[\w-]+/comments/[\w-]+/.*"
    ),
    re.compile(
        r"(?i)https?://redd\.it/[\w-]+"
    ),

    # Vimeo
    re.compile(
        r"(?i)https?://(?:www\.)?vimeo\.com/\d+"
    ),
]
