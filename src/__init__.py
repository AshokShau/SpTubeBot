import logging
from datetime import datetime

from pytdbot import Client, types

from src import config
from src.utils import HttpClient

logging.basicConfig(
    level=logging.INFO,
    format="[%(asctime)s - %(levelname)s] - %(name)s - %(filename)s:%(lineno)d - %(message)s",
    datefmt="%d-%b-%y %H:%M:%S",
    handlers=[logging.StreamHandler()],
)

logging.getLogger("httpx").setLevel(logging.WARNING)

LOGGER = logging.getLogger("Bot")

StartTime = datetime.now()


class Telegram(Client):
    def __init__(self) -> None:
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

    async def stop(self) -> None:
        await self._http_client.close_client()
        await super().stop()


client: Telegram = Telegram()
