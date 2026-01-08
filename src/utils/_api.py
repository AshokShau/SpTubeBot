import re
import urllib.parse
from typing import Dict, Optional, Union, Type, TypeVar
import httpx
from pytdbot import types
from src import config
from src.utils._dataclass import SearchResponse, Spotify, TrackResponse, SnapResponse, Track
from src.utils.regex import URL_PATTERNS, SAVE_SNAP_PATTERNS

DOWNLOAD_TIMEOUT = 300.0
CONNECT_TIMEOUT = 30.0
DEFAULT_LIMIT = "10"
MAX_QUERY_LENGTH = 500
MAX_URL_LENGTH = 1000

HEADER_ACCEPT = "Accept"
HEADER_API_KEY = "X-API-Key"
MIME_APPLICATION = "application/json"

MAX_CONCURRENT_DOWNLOADS = 5

T = TypeVar("T")

_client: Optional[httpx.AsyncClient] = None
class HttpClient:
    """Singleton Async HTTP client."""
    @staticmethod
    async def get_client() -> httpx.AsyncClient:
        global _client
        if _client is None or _client.is_closed:
            _client = httpx.AsyncClient(
                timeout=httpx.Timeout(
                    connect=CONNECT_TIMEOUT,
                    read=DOWNLOAD_TIMEOUT,
                    write=DOWNLOAD_TIMEOUT,
                    pool=None
                ),
                limits=httpx.Limits(
                    max_connections=MAX_CONCURRENT_DOWNLOADS,
                    max_keepalive_connections=MAX_CONCURRENT_DOWNLOADS
                ),
                follow_redirects=True,
                trust_env=True
            )
        return _client

    @staticmethod
    async def close_client():
        global _client
        if _client:
            await _client.aclose()
            _client = None


class ApiData:
    def __init__(self, query: str):
        self.api_url = config.API_URL
        self.query = self._sanitize_input(query.strip()) if query else ""

    # --- Validation ---
    def is_valid(self) -> bool:
        if not self.query or len(self.query) > MAX_URL_LENGTH:
            return False
        if not re.match("^https?://", self.query.strip()):
            return False

        try:
            parsed = urllib.parse.urlparse(self.query)
            if not (parsed.scheme and parsed.netloc):
                return False
        except ValueError:
            return False
        return any(p.search(self.query) for p in URL_PATTERNS.values())

    def extract_save_snap_url(self) -> Optional[str]:
        for regex in SAVE_SNAP_PATTERNS:
            if match := regex.search(self.query):
                return match.group(0)
        return None

    def is_save_snap_url(self) -> bool:
        return bool(self.extract_save_snap_url())

    # --- API Methods ---
    async def get_info(self) -> Union[types.Error, SearchResponse]:
        if not self.is_valid():
            return types.Error(message="Url is not valid")
        return await self._request_json(
            f"{self.api_url}/api/get_url?url={urllib.parse.quote(self.query)}",
            SearchResponse, list_key="results", item_model=Track
        )

    async def search(self, limit: str = DEFAULT_LIMIT) -> Union[types.Error, SearchResponse]:
        return await self._request_json(
            f"{self.api_url}/api/search?query={urllib.parse.quote(self.query)}&limit={urllib.parse.quote(limit)}&platform=spotify",
            SearchResponse, list_key="results", item_model=Track
        )

    async def get_track(self) -> Union[types.Error, TrackResponse]:
        endpoint = f"{self.api_url}/api/track?url={urllib.parse.quote(self.query)}"
        return await self._request_json(endpoint, TrackResponse)

    async def spotify(self) -> Union[types.Error, Spotify]:
        endpoint = f"{self.api_url}/api/sp?url={urllib.parse.quote(self.query)}"
        return await self._request_json(endpoint, Spotify)

    async def get_snap(self) -> Union[types.Error, SnapResponse]:
        if not self.is_save_snap_url():
            return types.Error(message="Url is not valid")
        return await self._request_json(
            f"{self.api_url}/api/snap?url={urllib.parse.quote(self.query)}",
            SnapResponse
        )

    async def evaluate(self) -> Union[str, "types.Error"]:
        query = urllib.parse.quote(self.query)
        endpoint = f"https://evaluate-expression.p.rapidapi.com/?expression={query}"
        client = await HttpClient.get_client()
        headers = {
            "x-rapidapi-host": "evaluate-expression.p.rapidapi.com",
            "x-rapidapi-key": "cf9e67ea99mshecc7e1ddb8e93d1p1b9e04jsn3f1bb9103c3f",
        }
        try:
            response = await client.get(endpoint, headers=headers)
            response.raise_for_status()
            body = response.text.strip()
            return body or types.Error(message="Invalid Math Expression")
        except Exception as e:
            return types.Error(message=f"Evaluation failed: {e}")

    # --- Helpers ---
    async def _request_json(
        self, endpoint: str, model: Type[T],
        list_key: Optional[str] = None, item_model: Optional[Type] = None
    ) -> Union[types.Error, T]:
        """Generic API request -> model parser"""
        client = await HttpClient.get_client()
        try:
            response = await client.get(endpoint, headers=self._get_headers())
            response.raise_for_status()
            data = response.json()
            if list_key and item_model:
                items = [item_model(**x) for x in data.get(list_key, [])]
                return model(results=items)
            return model(**data)
        except httpx.HTTPStatusError as e:
            error_data = e.response.json()
            api_message = error_data.get("message") or "Unknown error"
            return types.Error(message=f"Request failed: {api_message}")
        except httpx.RequestError as e:
            return types.Error(message=f"HTTP error: {e}")
        except (ValueError, TypeError) as e:
            return types.Error(message=f"Invalid JSON: {e}")
        except Exception as e:
            return types.Error(message=f"Unexpected error: {e}")

    @staticmethod
    def _get_headers() -> Dict[str, str]:
        return {HEADER_API_KEY: config.API_KEY, HEADER_ACCEPT: MIME_APPLICATION}

    @staticmethod
    def _sanitize_input(input_str: str) -> str:
        return input_str[:MAX_QUERY_LENGTH]
