import re
from typing import Union

from pytdbot import filters, types


class Filter:
    @staticmethod
    def _extract_text(event) -> str | None:
        """
        Extract text from an event. If the event is a message, it will get the text from
        the message. If the event is an update, it will get the text from the message.
        If the event is a callback query, it will decode the data using UTF-8 and return
        the result. If it can't extract the text, it will return None.

        :param event: The event to extract the text from.
        :return: The text extracted from the event, or None if the text couldn't be
            extracted.
        """
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
