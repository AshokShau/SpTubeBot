import re
import urllib.parse
from typing import Dict, Union, Optional

import aiohttp
from pytdbot import types

from src import config
from src.utils._dataclass import PlatformTracks, TrackInfo, MusicTrack

# Constants
API_TIMEOUT = aiohttp.ClientTimeout(total=60, connect=10)
DEFAULT_LIMIT = "10"
MAX_QUERY_LENGTH = 500
MAX_URL_LENGTH = 5000
HEADER_ACCEPT = "Accept"
HEADER_API_KEY = "X-API-Key"
MIME_APPLICATION = "application/json"

_client_session: Optional[aiohttp.ClientSession] = None

URL_PATTERNS = {
    "spotify": re.compile(
        r'^(https?://)?([a-z0-9-]+\.)*spotify\.com/(track|playlist|album|artist)/[a-zA-Z0-9]+(\?.*)?$'),
    "youtube": re.compile(r'^(https?://)?([a-z0-9-]+\.)*(youtube\.com/watch\?v=|youtu\.be/)[\w-]+(\?.*)?$'),
    "youtube_music": re.compile(r'^(https?://)?([a-z0-9-]+\.)*youtube\.com/(watch\?v=|playlist\?list=)[\w-]+(\?.*)?$'),
    "soundcloud": re.compile(r'^(https?://)?([a-z0-9-]+\.)*soundcloud\.com/[\w-]+(/[\w-]+)?(/sets/[\w-]+)?(\?.*)?$'),
    "apple_music": re.compile(
        r'^(https?://)?([a-z0-9-]+\.)?apple\.com/[a-z]{2}/(album|playlist|song)/[^/]+/(pl\.[a-zA-Z0-9]+|\d+)(\?i=\d+)?(\?.*)?$')
}


async def get_client_session() -> aiohttp.ClientSession:
    """Get or create a reusable client session."""
    global _client_session
    if _client_session is None or _client_session.closed:
        connector = aiohttp.TCPConnector(
            limit=100,
            force_close=False,
            enable_cleanup_closed=True,
            use_dns_cache=True
        )
        _client_session = aiohttp.ClientSession(
            connector=connector,
            timeout=API_TIMEOUT,
            trust_env=True
        )
    return _client_session


class ApiData:
    def __init__(self, query: str):
        self.api_url = config.API_URL
        self.query = self._sanitize_input(query)

    def is_valid(self) -> bool:
        raw_url = self.query
        if not raw_url or len(raw_url) > MAX_URL_LENGTH:
            return False

        if not (raw_url.startswith('http://') or raw_url.startswith('https://')):
            return False

        try:
            parsed = urllib.parse.urlparse(raw_url)
            if not all([parsed.scheme, parsed.netloc]):
                return False
        except ValueError:
            return False

        return any(pattern.search(raw_url) for pattern in URL_PATTERNS.values())

    async def get_info(self) -> Union[types.Error, PlatformTracks]:
        if not self.is_valid():
            return types.Error(message="Url is not valid")
        return await self._fetch_data(self.query)

    async def _fetch_data(self, raw_url: str) -> Union[types.Error, PlatformTracks]:
        endpoint = f"{self.api_url}/get_url?url={urllib.parse.quote(raw_url)}"
        headers = self._get_headers()
        session = await get_client_session()

        try:
            async with session.get(
                    endpoint,
                    headers=headers,
                    raise_for_status=True
            ) as response:
                raw_data = await response.json(content_type=None)
                results = [MusicTrack(**track) for track in raw_data.get("results", [])]
                return PlatformTracks(results=results)
        except aiohttp.ClientResponseError as e:
            return types.Error(message=f"Request failed with status: {e.status}")
        except aiohttp.ClientError as e:
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
        session = await get_client_session()

        try:
            async with session.get(
                    endpoint,
                    headers=headers,
                    raise_for_status=True
            ) as response:
                raw_data = await response.json(content_type=None)
                results = [MusicTrack(**track) for track in raw_data.get("results", [])]
                return PlatformTracks(results=results)
        except aiohttp.ClientResponseError as e:
            return types.Error(message=f"Request failed with status: {e.status}")
        except aiohttp.ClientError as e:
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
        session = await get_client_session()

        try:
            async with session.get(
                    endpoint,
                    headers=headers,
                    raise_for_status=True
            ) as response:
                raw_data = await response.json(content_type=None)
                return TrackInfo(**raw_data)

        except aiohttp.ClientResponseError as e:
            return types.Error(message=f"Request failed with status: {e.status}")
        except aiohttp.ClientError as e:
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


async def close_client_session():
    """Close the global client session when done."""
    global _client_session
    if _client_session is not None and not _client_session.closed:
        await _client_session.close()
