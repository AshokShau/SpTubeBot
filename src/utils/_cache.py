import hashlib
from threading import RLock
from typing import Dict, Optional


class URLShortener:
    def __init__(self):
        self.url_map: Dict[str, str] = {}
        self._lock = RLock()
        self._token_len = 10

    def encode_url(self, url: str) -> str:
        """Stores the URL and returns a short token (default 10 characters)"""
        token = self._generate_short_token(url)

        with self._lock:
            self.url_map[token] = url

        return token

    def decode_url(self, token: str) -> Optional[str]:
        """Retrieves the original URL using the token"""
        with self._lock:
            return self.url_map.get(token)

    def _generate_short_token(self, url: str) -> str:
        """Creates a consistent short hash from the URL"""
        # Create SHA-256 hash
        hash_obj = hashlib.sha256(url.encode())
        hex_digest = hash_obj.hexdigest()

        # Return first N characters of the hex digest
        return hex_digest[:self._token_len]

shortener = URLShortener()
