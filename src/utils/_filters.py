import re
from typing import Union

from pytdbot import filters, types

from src.utils._api import ApiData


class Filter:
    @staticmethod
    def _extract_text(event) -> str | None:
        if isinstance(event, types.Message) and isinstance(
            event.content, types.MessageText
        ):
            return event.content.text.text
        if isinstance(event, types.UpdateNewMessage) and isinstance(
            event.message, types.MessageText
        ):
            return event.message.text.text
        if isinstance(event, types.UpdateNewCallbackQuery) and event.payload:
            return event.payload.data.decode()

        return None

    @staticmethod
    def command(
        commands: Union[str, list[str]], prefixes: str = "/!"
    ) -> filters.Filter:
        """
        Filter for commands.

        Supports multiple commands and prefixes like / or !. Also handles commands with
        @mentions (e.g., /start@BotName).
        """
        if isinstance(commands, str):
            commands = [commands]
        commands_set = {cmd.lower() for cmd in commands}

        pattern = re.compile(
            rf"^[{re.escape(prefixes)}](\w+)(?:@(\w+))?", re.IGNORECASE
        )

        async def filter_func(client, event) -> bool:
            text = Filter._extract_text(event)
            if not text:
                return False

            match = pattern.match(text.strip())
            if not match:
                return False

            cmd, mentioned_bot = match.groups()
            if cmd.lower() not in commands_set:
                return False

            if mentioned_bot:
                bot_username = client.me.usernames.editable_username
                return bot_username and mentioned_bot.lower() == bot_username.lower()

            return True

        return filters.create(filter_func)

    @staticmethod
    def regex(pattern: str) -> filters.Filter:
        """
        Filter for messages or callback queries matching a regex pattern.
        """

        compiled = re.compile(pattern)

        async def filter_func(_, event) -> bool:
            text = Filter._extract_text(event)
            return bool(compiled.search(text)) if text else False

        return filters.create(filter_func)

    @staticmethod
    def save_snap() -> filters.Filter:
        insta_regex = re.compile(r"(?i)https?://(?:www\.)?(instagram\.com|instagr\.am)/(reel|stories|p|tv)/[^\s/?]+")
        pin_regex = re.compile(r"(?i)https?://(?:[a-z]+\.)?(pinterest\.com|pin\.it)/[^\s]+")
        fb_watch_regex = re.compile(r"(?i)https?://(?:www\.)?fb\.watch/[^\s/?]+")
        fb_video_regex = re.compile(r"(?i)https?://(?:www\.)?facebook\.com/.+/videos/\d+")

        async def filter_func(_, event) -> bool:
            text = Filter._extract_text(event)
            if not text:
                return False

            prefixes: str = "/!"
            pattern = re.compile(rf"^[{re.escape(prefixes)}](\w+)(?:@(\w+))?", re.IGNORECASE)
            if pattern.match(text.strip()):
                return False

            return any(
                regex.search(text)
                for regex in (insta_regex, pin_regex, fb_watch_regex, fb_video_regex)
            )

        return filters.create(filter_func)

    @staticmethod
    def sp_tube() -> filters.Filter:
        async def filter_func(_, event) -> bool:
            text = Filter._extract_text(event)
            if not text:
                return False

            # Skip command-like messages
            command_pattern = re.compile(r"^[!/](\w+)(?:@\w+)?", re.IGNORECASE)
            if command_pattern.match(text.strip()):
                return False

            chat_id = None
            if isinstance(event, types.Message):
                chat_id = event.chat_id
            elif isinstance(event, types.UpdateNewMessage):
                chat_id = getattr(event.message, "chat_id", None)

            if not chat_id or chat_id <= 0:
                return False

            return ApiData(text).is_valid()

        return filters.create(filter_func)
