import re
from typing import Union, Optional, Pattern, Set

from pytdbot import filters, types
from src.utils._api import ApiData


class Filter:
    @staticmethod
    def _extract_text(event: Union[types.Message, types.UpdateNewMessage, types.UpdateNewCallbackQuery]) -> Optional[
        str]:
        """Extract text content from different event types."""
        if isinstance(event, types.Message) and isinstance(event.content, types.MessageText):
            return event.content.text.text
        if isinstance(event, types.UpdateNewMessage) and isinstance(event.message, types.MessageText):
            return event.message.text.text
        if isinstance(event, types.UpdateNewCallbackQuery) and event.payload:
            return event.payload.data.decode()
        return None

    @staticmethod
    def command(commands: Union[str, list[str]], prefixes: str = "/!") -> filters.Filter:
        """
        Filter for commands.

        Args:
            commands: Single command or list of commands to match
            prefixes: String of allowed command prefixes (default: "/!")

        Returns:
            Filter that matches commands with optional bot username mention
        """
        commands_set: Set[str] = {cmd.lower() for cmd in ([commands] if isinstance(commands, str) else commands)}
        pattern: Pattern = re.compile(
            rf"^[{re.escape(prefixes)}](\w+)(?:@(\w+))?",
            re.IGNORECASE
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
    def regex(pattern: str, flags: int = 0) -> filters.Filter:
        """
        Filter for messages or callback queries matching a regex pattern.

        Args:
            pattern: Regex pattern to match
            flags: Regex flags (default: 0)

        Returns:
            Filter that matches text against the compiled regex pattern
        """
        compiled: Pattern = re.compile(pattern, flags)

        async def filter_func(_, event) -> bool:
            text = Filter._extract_text(event)
            return bool(compiled.search(text)) if text else False

        return filters.create(filter_func)

    @staticmethod
    def save_snap() -> filters.Filter:
        """
        Filter for Snapchat URLs that should be saved.

        Returns:
            Filter that matches valid Snapchat URLs and excludes commands
        """
        command_pattern: Pattern = re.compile(r"^[!/]\w+(?:@\w+)?", re.IGNORECASE)

        async def filter_func(_, event) -> bool:
            text = Filter._extract_text(event)
            if not text or command_pattern.match(text.strip()):
                return False
            return ApiData(text).is_save_snap_url()

        return filters.create(filter_func)

    @staticmethod
    def sp_tube() -> filters.Filter:
        """
        Filter for special tube URLs with additional checks.

        Returns:
            Filter that matches valid URLs and handles private/group chat logic
        """
        command_pattern: Pattern = re.compile(r"^[!/]\w+(?:@\w+)?", re.IGNORECASE)

        async def filter_func(client, event) -> bool:
            text = Filter._extract_text(event)
            if not text or command_pattern.match(text.strip()):
                return False

            chat_id: Optional[int] = None
            if isinstance(event, types.Message):
                if event.via_bot_user_id == client.me.id:
                    return False
                chat_id = event.chat_id
            elif isinstance(event, types.UpdateNewMessage):
                if event.message.via_bot_user_id == client.me.id:
                    return False
                chat_id = getattr(event.message, "chat_id", None)

            if chat_id is None:
                return False

            if ApiData(text).is_valid():
                return True

            if re.match("^https?://", text):
                return False

            return chat_id > 0

        return filters.create(filter_func)
