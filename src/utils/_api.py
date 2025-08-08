import re
import urllib.parse
from typing import Dict, Union, Optional
import httpx
from pytdbot import types
from concurrent.futures import ThreadPoolExecutor
from src import config
from src.utils._dataclass import PlatformTracks, TrackInfo, MusicTrack, APIResponse

# Constants
DOWNLOAD_TIMEOUT = 300.0  # Total timeout in seconds
CONNECT_TIMEOUT = 30.0  # Connect timeout in seconds
DEFAULT_LIMIT = "10"
MAX_QUERY_LENGTH = 500
MAX_URL_LENGTH = 5000
HEADER_ACCEPT = "Accept"
HEADER_API_KEY = "X-API-Key"
MIME_APPLICATION = "application/json"
MAX_CONCURRENT_DOWNLOADS = 5
_client: Optional[httpx.AsyncClient] = None
_executor = ThreadPoolExecutor(max_workers=4)

URL_PATTERNS = {
    "spotify": re.compile(
        r'^(https?://)?([a-z0-9-]+\.)*spotify\.com/(track|playlist|album|artist)/[a-zA-Z0-9]+(\?.*)?$'),
    "youtube": re.compile(r'^(https?://)?([a-z0-9-]+\.)*(youtube\.com/watch\?v=|youtu\.be/)[\w-]+(\?.*)?$'),
    "youtube_music": re.compile(r'^(https?://)?([a-z0-9-]+\.)*youtube\.com/(watch\?v=|playlist\?list=)[\w-]+(\?.*)?$'),
    "soundcloud": re.compile(r'^(https?://)?([a-z0-9-]+\.)*soundcloud\.com/[\w-]+(/[\w-]+)?(/sets/[\w-]+)?(\?.*)?$'),
    "apple_music": re.compile(
        r'^(https?://)?([a-z0-9-]+\.)?apple\.com/[a-z]{2}/(album|playlist|song)/[^/]+/(pl\.[a-zA-Z0-9]+|\d+)(\?i=\d+)?(\?.*)?$')
}


class HttpClient:
    @staticmethod
    async def get_client() -> httpx.AsyncClient:
        global _client
        if _client is None or _client.is_closed:
            limits = httpx.Limits(
                max_connections=MAX_CONCURRENT_DOWNLOADS,
                max_keepalive_connections=MAX_CONCURRENT_DOWNLOADS
            )

            _client = httpx.AsyncClient(
                timeout=httpx.Timeout(
                    connect=CONNECT_TIMEOUT,
                    read=DOWNLOAD_TIMEOUT,
                    write=DOWNLOAD_TIMEOUT,
                    pool=None
                ),
                limits=limits,
                follow_redirects=True,
                trust_env=True
            )
        return _client

    @staticmethod
    async def close_client():
        global _client
        if _client is not None:
            await _client.aclose()
            _client = None


class ApiData:
    def __init__(self, query: str):
        self.api_url = config.API_URL
        self.query = self._sanitize_input(query) if query else ""

    def is_valid(self) -> bool:
        raw_url = self.query
        if not raw_url or len(raw_url) > MAX_URL_LENGTH:
            return False

        if not re.match("^https?://", raw_url):
            return False

        try:
            parsed = urllib.parse.urlparse(raw_url)
            if not all([parsed.scheme, parsed.netloc]):
                return False
        except ValueError:
            return False

        return any(pattern.search(raw_url) for pattern in URL_PATTERNS.values())

    def is_save_snap_url(self) -> bool:
        insta_regex = re.compile(r"(?i)https?://(?:www\.)?(instagram\.com|instagr\.am)/(reel|stories|p|tv)/[^\s/?]+")
        pin_regex = re.compile(r"(?i)https?://(?:[a-z]+\.)?(pinterest\.com|pin\.it)/[^\s]+")
        fb_watch_regex = re.compile(r"(?i)https?://(?:www\.)?fb\.watch/[^\s/?]+")
        fb_video_regex = re.compile(r"(?i)https?://(?:www\.)?facebook\.com/.+/videos/\d+")
        tiktok_regex = re.compile(
            r"https?://(?:www\.|m\.)?(?:vt\.)?tiktok\.com/(?:@[\w.-]+/video/\d+|v/\d+\.html|t/[\w]+|[\w]+)",
            re.IGNORECASE
        )
        x_regex = re.compile(
            r"(https?://(?:www\.)?(?:x|twitter)\.com/[^\s]+)",
            re.IGNORECASE
        )
        threads_redex = re.compile(
            r'^https?://(?:www\.)?threads\.(?:com|net)/@[\w.-]+/post/[\w-]+(?:\?[\w=&%-]+)?$',
            re.IGNORECASE
        )
        reddit_regex = re.compile(
            r"https?://(?:www\.|old\.)?reddit\.com/r/[\w]+/comments/[\w]+(?:/[^\s]*)?|https?://redd\.it/[\w]+",
            re.IGNORECASE
        )

        return any(
            regex.search(self.query)
            for regex in (insta_regex, pin_regex, fb_watch_regex, fb_video_regex, tiktok_regex, x_regex, threads_redex, reddit_regex)
        )

    async def get_info(self) -> Union[types.Error, PlatformTracks]:
        if not self.is_valid():
            return types.Error(message="Url is not valid")
        return await self._fetch_data(self.query)

    async def _fetch_data(self, raw_url: str) -> Union[types.Error, PlatformTracks]:
        endpoint = f"{self.api_url}/get_url?url={urllib.parse.quote(raw_url)}"
        headers = self._get_headers()
        client = await HttpClient.get_client()

        try:
            response = await client.get(
                endpoint,
                headers=headers
            )
            response.raise_for_status()
            raw_data = response.json()
            results = [MusicTrack(**track) for track in raw_data.get("results", [])]
            return PlatformTracks(results=results)
        except httpx.HTTPStatusError as e:
            return types.Error(message=f"Request failed with status: {e.response.status_code}")
        except httpx.RequestError as e:
            return types.Error(message=f"HTTP request failed: {e}")
        except (ValueError, TypeError) as e:
            return types.Error(message=f"Failed to parse JSON response: {e}")
        except Exception as e:
            return types.Error(message=f"Unexpected error: {e}")

    async def search(self, limit: str = DEFAULT_LIMIT) -> Union[types.Error, PlatformTracks]:
        endpoint = (
            f"{self.api_url}/search_track/{urllib.parse.quote(self.query)}"
            f"?lim={urllib.parse.quote(limit)}"
        )
        headers = self._get_headers()
        client = await HttpClient.get_client()

        try:
            response = await client.get(
                endpoint,
                headers=headers
            )
            response.raise_for_status()
            raw_data = response.json()
            results = [MusicTrack(**track) for track in raw_data.get("results", [])]
            return PlatformTracks(results=results)
        except httpx.HTTPStatusError as e:
            return types.Error(message=f"Request failed with status: {e.response.status_code}")
        except httpx.RequestError as e:
            return types.Error(message=f"HTTP request failed: {e}")
        except (ValueError, TypeError) as e:
            return types.Error(message=f"Failed to parse JSON response: {e}")
        except Exception as e:
            return types.Error(message=f"Unexpected error: {e}")

    async def get_track(self) -> Union[types.Error, TrackInfo]:
        track_id = self.query
        if not track_id:
            return types.Error(message="Empty track ID")

        endpoint = f"{self.api_url}/get_track?id={urllib.parse.quote(track_id)}"
        headers = self._get_headers()
        client = await HttpClient.get_client()

        try:
            response = await client.get(
                endpoint,
                headers=headers
            )
            response.raise_for_status()
            raw_data = response.json()
            return TrackInfo(**raw_data)
        except httpx.HTTPStatusError as e:
            return types.Error(message=f"Request failed with status: {e.response.status_code}")
        except httpx.RequestError as e:
            return types.Error(message=f"HTTP request failed: {e}")
        except (ValueError, TypeError) as e:
            return types.Error(message=f"Failed to parse JSON response: {e}")
        except Exception as e:
            return types.Error(message=f"Unexpected error: {e}")

    async def get_snap(self) -> Union[types.Error, APIResponse]:
        if not self.is_save_snap_url():
            return types.Error(message="Url is not valid")

        endpoint = f"{self.api_url}/snap?url={urllib.parse.quote(self.query)}"
        headers = self._get_headers()
        client = await HttpClient.get_client()

        try:
            response = await client.get(
                endpoint,
                headers=headers
            )
            response.raise_for_status()
            raw_data = response.json()
            return APIResponse(**raw_data)
        except httpx.HTTPStatusError as e:
            return types.Error(message=f"Request failed with status: {e.response.status_code}")
        except httpx.RequestError as e:
            return types.Error(message=f"HTTP request failed: {e}")
        except (ValueError, TypeError) as e:
            return types.Error(message=f"Failed to parse JSON response: {e}")
        except Exception as e:
            return types.Error(message=f"Unexpected error: {e}")

    @staticmethod
    def _get_headers() -> Dict[str, str]:
        return {
            HEADER_API_KEY: config.API_KEY,
            HEADER_ACCEPT: MIME_APPLICATION
        }

    @staticmethod
    def _sanitize_input(input_str: str) -> str:
        return input_str[:MAX_QUERY_LENGTH] if len(input_str) > MAX_QUERY_LENGTH else input_str
