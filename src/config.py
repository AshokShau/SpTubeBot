from os import getenv
from pathlib import Path
from typing import Optional

from dotenv import load_dotenv

load_dotenv()


def get_env_int(name: str, default: Optional[int] = None) -> Optional[int]:
    value = getenv(name)
    try:
        return int(value)
    except (TypeError, ValueError):
        print(f"Invalid value for {name}: {value}")
        return default


API_ID: Optional[int] = get_env_int("API_ID")
API_HASH: Optional[str] = getenv("API_HASH")
TOKEN: Optional[str] = getenv("TOKEN")
API_KEY = getenv("API_KEY")
API_URL = getenv("API_URL")
DOWNLOAD_PATH = Path(getenv("DOWNLOAD_PATH", "database/music"))
MONGO_URI: Optional[str] = getenv("MONGO_URI")
LOGGER_ID = get_env_int("LOGGER_ID", -1002434755494)
COOKIES_URL: Optional[str] = getenv("COOKIES_URL")
