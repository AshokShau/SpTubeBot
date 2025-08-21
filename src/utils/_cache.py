import hashlib
from typing import Dict, Optional


class URLShortener:
    def __init__(self) -> None:
        self.url_map: Dict[str, str] = {}
        self._token_len: int = 10

    def encode_url(self, url: str) -> str:
        token = self._generate_short_token(url)
        self.url_map[token] = url
        return token

    def decode_url(self, token: str) -> Optional[str]:
        return self.url_map.get(token)

    def _generate_short_token(self, url: str) -> str:
        hash_obj = hashlib.sha256(url.encode())
        return hash_obj.hexdigest()[:self._token_len]


shortener = URLShortener()
