from os import getenv
from typing import Optional

from dotenv import load_dotenv

load_dotenv()


def get_env_int(name: str, default: Optional[int] = None) -> Optional[int]:
    value = getenv(name)
    try:
        return int(value)
    except (TypeError, ValueError):
        return default


API_ID: Optional[int] = get_env_int("API_ID")
API_HASH: Optional[str] = getenv("API_HASH")
TOKEN: Optional[str] = getenv("TOKEN")
API_KEY = getenv("API_KEY")
API_URL = getenv("API_URL")
DOWNLOAD_PATH = getenv("DOWNLOAD_PATH", "database")
LOGGER_ID = get_env_int("LOGGER_ID", -1002434755494)

# MongoDB configuration
MONGO_URI: Optional[str] = getenv("MONGO_URI")
MONGO_DB_NAME: str = getenv("MONGO_DB_NAME", "sptubebot")
MONGO_UPLOADS_COLL: str = getenv("MONGO_UPLOADS_COLL", "uploads_cache")
