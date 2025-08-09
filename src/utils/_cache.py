import hashlib
import logging
from collections import OrderedDict
from datetime import datetime
from threading import RLock
from typing import Dict, Optional

from src import config

logger = logging.getLogger(__name__)


class InMemoryUploadCache:
    def __init__(self, max_items: int = 2000):
        self.cache = OrderedDict()  # key -> file_id
        self.meta: Dict[str, Dict[str, object]] = {}  # key -> refs
        self.max_items = max_items

    def get(self, key: str) -> Optional[str]:
        return self.cache.get(key)

    def get_message_ref(self, key: str) -> Optional[Dict[str, object]]:
        return self.meta.get(key)

    def set(
        self,
        key: str,
        file_id: str,
        message_link: Optional[str] = None,
        chat_id: Optional[int] = None,
        message_id: Optional[int] = None,
    ) -> None:
        self.cache[key] = file_id
        if message_link or chat_id or message_id:
            self.meta[key] = {
                "message_link": message_link,
                "chat_id": chat_id,
                "message_id": message_id,
            }
        self.cache.move_to_end(key)
        if len(self.cache) > self.max_items:
            # also drop meta for evicted key if exists
            evicted_key, _ = self.cache.popitem(last=False)
            self.meta.pop(evicted_key, None)


class InMemoryURLShortener:
    def __init__(self, token_length: int = 10):
        self.url_map: Dict[str, str] = {}
        self._lock = RLock()
        self._token_len = token_length

    def encode_url(self, url: str) -> str:
        token = self._generate_short_token(url)
        with self._lock:
            self.url_map[token] = url
        return token

    def decode_url(self, token: str) -> Optional[str]:
        with self._lock:
            return self.url_map.get(token)

    def _generate_short_token(self, url: str) -> str:
        hash_obj = hashlib.sha256(url.encode())
        hex_digest = hash_obj.hexdigest()
        return hex_digest[:self._token_len]


class MongoUploadCache:
    def __init__(self):
        from pymongo import ASCENDING, MongoClient

        if not config.MONGO_URI:
            raise RuntimeError("MONGO_URI is not set")

        client = MongoClient(config.MONGO_URI)
        db = client[config.MONGO_DB_NAME]
        self.col = db[config.MONGO_UPLOADS_COLL]
        # Ensure indexes: use _id as the unique identifier
        # No extra indexes required beyond default _id

        # Keep in-memory LRU as a tiny hot cache to reduce roundtrips
        self.hot_cache = InMemoryUploadCache(max_items=512)

    def get(self, key: str) -> Optional[str]:
        hot = self.hot_cache.get(key)
        if hot is not None:
            return hot
        try:
            doc = self.col.find_one({"_id": key}, {"_id": 0, "file_id": 1})
            file_id = doc.get("file_id") if doc else None
            if file_id:
                self.hot_cache.set(key, file_id)
            logger.info(f"MongoUploadCache.get: _id={key} hit={'yes' if file_id else 'no'}")
            return file_id
        except Exception as e:
            logger.warning(f"Mongo get failed for key={key}: {e}")
            return None

    def get_message_ref(self, key: str) -> Optional[Dict[str, object]]:
        try:
            doc = self.col.find_one({"_id": key}, {"_id": 0, "msg_url": 1})
            if not doc:
                return None
            return {"message_link": doc.get("msg_url")}
        except Exception as e:
            logger.warning(f"Mongo get_message_ref failed for key={key}: {e}")
            return None

    def set(
        self,
        key: str,
        file_id: str,
        message_link: Optional[str] = None,
        chat_id: Optional[int] = None,
        message_id: Optional[int] = None,
    ) -> None:
        try:
            # Only persist minimal fields: _id (track id), file_id, msg_url
            self.col.update_one(
                {"_id": key},
                {"$set": {"file_id": file_id, "msg_url": message_link}},
                upsert=True,
            )
            logger.info(f"MongoUploadCache.set: saved _id={key} file_id={'set' if bool(file_id) else 'missing'} msg_url={'set' if bool(message_link) else 'missing'}")
        except Exception as e:
            logger.warning(f"Mongo set failed for key={key}: {e}")
        finally:
            self.hot_cache.set(key, file_id)


class MongoURLShortener:
    def __init__(self, token_length: int = 10):
        # Per schema baru: tidak menyimpan short URL di Mongo.
        self._in_memory = InMemoryURLShortener(token_length=token_length)

    def encode_url(self, url: str) -> str:
        return self._in_memory.encode_url(url)

    def decode_url(self, token: str) -> Optional[str]:
        return self._in_memory.decode_url(token)


def _build_instances():
    """Build cache/shortener instances with Mongo if available, else in-memory."""
    # Default fallbacks
    upload: InMemoryUploadCache | MongoUploadCache
    short: InMemoryURLShortener | MongoURLShortener

    try:
        if config.MONGO_URI:
            # Delay import so environments without pymongo still run
            import pymongo  # noqa: F401
            upload = MongoUploadCache()
            short = MongoURLShortener()
            logger.info("Using MongoDB for upload cache; shortener is in-memory")
        else:
            raise RuntimeError("MONGO_URI not configured")
    except Exception as e:
        logger.warning(f"Falling back to in-memory caches: {e}")
        upload = InMemoryUploadCache()
        short = InMemoryURLShortener()

    return upload, short


upload_cache, shortener = _build_instances()
