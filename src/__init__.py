import logging
from datetime import datetime

from pytdbot import Client, types

from src import config

logging.basicConfig(
    level=logging.INFO,
    format="[%(asctime)s - %(levelname)s] - %(name)s - %(filename)s:%(lineno)d - %(message)s",
    datefmt="%d-%b-%y %H:%M:%S",
    handlers=[logging.StreamHandler()],
)


LOGGER = logging.getLogger("Bot")

__version__ = "0.1.0"
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

    async def start(self) -> None:
        await super().start()
        self.logger.info(f"Bot started in {datetime.now() - StartTime} seconds.")
        self.logger.info(f"Version: {__version__}")

    async def stop(self) -> None:
        await super().stop()
        from src.utils import close_client_session
        await close_client_session()


client: Telegram = Telegram()
