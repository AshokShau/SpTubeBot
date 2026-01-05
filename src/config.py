from decouple import config
from pathlib import Path
from typing import Optional

API_ID: Optional[int] = config("API_ID", default=6, cast=int)
API_HASH: Optional[str] = config("API_HASH", default="", cast=str)
TOKEN: Optional[str] = config("TOKEN", default="")
API_KEY: str = config("API_KEY")
API_URL: str = config("API_URL", default="https://beta.fallenapi.fun", cast=str)
DOWNLOAD_PATH: Path = Path(config("DOWNLOAD_PATH", "database/music"))
MONGO_URI: Optional[str] = config("MONGO_URI", default="")
LOGGER_ID: int = config("LOGGER_ID", default=-1002434755494, cast=int)
FSUB_ID: int = config("FSUB_ID", default=0, cast=int)
OWNER_ID: int = config("OWNER_ID", default=5938660179, cast=int)
