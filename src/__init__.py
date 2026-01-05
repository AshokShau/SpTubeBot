import logging
import shutil
from datetime import datetime

from pytdbot import Client, types

from src import config
from src.utils import HttpClient, db

logging.basicConfig(
    level=logging.INFO,
    format="[%(asctime)s - %(levelname)s] - %(name)s - %(filename)s:%(lineno)d - %(message)s",
    datefmt="%d-%b-%y %H:%M:%S",
    handlers=[logging.StreamHandler()],
)

logging.getLogger("httpx").setLevel(logging.WARNING)
logging.getLogger("pymongo").setLevel(logging.WARNING)

LOGGER = logging.getLogger("Bot")
StartTime = datetime.now()


class Telegram(Client):
    def __init__(self) -> None:
        self._check_config()
        super().__init__(
            token=config.TOKEN,
            api_id=config.API_ID,
            api_hash=config.API_HASH,
            default_parse_mode="html",
            td_verbosity=2,
            td_log=types.LogStreamEmpty(),
            plugins=types.plugins.Plugins(folder="src/modules"),
            files_directory="",
            database_encryption_key="",
            options={"ignore_background_updates": True},
        )

        self._http_client = HttpClient()

    async def start(self) -> None:
        await self._http_client.get_client()
        await super().start()
        self.logger.info(f"Bot started in {datetime.now() - StartTime} seconds.")
        await db.connect()

    async def stop(self) -> None:
        await self._http_client.close_client()
        await db.close()
        await super().stop()

    @staticmethod
    def _check_config() -> None:
        if not shutil.which('ffmpeg'):
            raise RuntimeError(
                "FFmpeg is not installed or not in system PATH. "
                "Please install FFmpeg to run this bot."
            )
            
        required_keys = [
            "TOKEN",
            "API_ID",
            "API_HASH",
            "API_KEY",
            "API_URL",
            "MONGO_URI",
            "LOGGER_ID",
        ]

        if missing := [
            key for key in required_keys if not getattr(config, key, None)
        ]:
            raise RuntimeError(
                f"Missing required config values in .env: {', '.join(missing)}"
            )

client: Telegram = Telegram()
