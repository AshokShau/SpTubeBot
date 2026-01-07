import re
from typing import Tuple, Optional, Union, TYPE_CHECKING
from pytdbot import types, Client
from pytdbot.types import Error, InputFileLocal, InputFileRemote

from src.utils import ApiData, Download, shortener, db

if TYPE_CHECKING:
    from src.utils._dataclass import Track, Spotify, TrackResponse


async def process_track_media(c: Client, track:  'TrackResponse', chat_id: Optional[int] = None, message_id: Optional[int] = None, inline_message_id: Optional[str] = None) -> Error | tuple[InputFileLocal | InputFileRemote, str | None]:
    parsed_status = await c.parseTextEntities("<b>Processing your track, please wait...</b>", types.TextParseModeHTML())
    text = types.InputMessageText(parsed_status)
    # Update status message
    if inline_message_id:
        await c.editInlineMessageText(inline_message_id=inline_message_id, input_message_content=text)
    elif chat_id and message_id:
        await c.editMessageText(chat_id=chat_id, message_id=message_id, input_message_content=text)

    audio_file, cover = None, None
    audio: Optional[Union[types.InputFileLocal, types.InputFileRemote]] = None
    api = ApiData(track.url)

    if track.platform.lower() == "spotify":
        _track = await api.spotify()
        if isinstance(_track, types.Error):
            error_msg = f"Download failed.\n<b>{_track.message}</b>"
            return types.Error(message=error_msg)

        dl = Download(_track)
        result = await dl.process()
        if isinstance(result, types.Error):
            error_msg = f"❌ Download failed.\n<b>{result.message}</b>"
            return types.Error(message=error_msg)

        audio_file, cover = result
        if not audio_file:
            return types.Error(message="Failed to download song.\nPlease report this to @FallenProjects.")

        file_id = await db.upload_song_and_get_file_id(audio_file, cover, _track)
        if isinstance(file_id, types.Error):
            return types.Error(message=file_id.message or "❌ Failed to send song to database.")

        audio = types.InputFileRemote(file_id)
        return audio, cover


    dl = Download(track)
    result = await dl.process()
    if isinstance(result, types.Error):
        error_msg = f"❌ Download failed.\n<b>{result.message}</b>"
        return types.Error(message=error_msg)

    audio_file, cover = result
    if not audio_file:
        error_msg = "❌ Failed to download song.\nPlease report this to @FallenProjects."
        return types.Error(message=error_msg)

    if re.match(r"https?://t\.me/([^/]+)/(\d+)", audio_file):
        info = await c.getMessageLinkInfo(audio_file)
        if isinstance(info, types.Error) or not info.message:
            return types.Error(message=f"❌ Failed to resolve link: {audio_file}")

        public_msg = await c.getMessage(info.chat_id, info.message.id)
        if isinstance(public_msg, types.Error):
            return types.Error(message="❌ Failed to fetch message: {public_msg.message}")

        if isinstance(public_msg.content, types.MessageAudio):
            audio = types.InputFileRemote(public_msg.content.audio.audio.remote.id)
        elif isinstance(public_msg.content, types.MessageDocument):
            audio = types.InputFileRemote(public_msg.content.document.document.remote.id)
        elif isinstance(public_msg.content, types.MessageVideo):
            audio = types.InputFileRemote(public_msg.content.video.video.remote.id)
        else:
            return types.Error(message=f"No audio file in t.me link: {audio_file}")
    else:
        audio = types.InputFileLocal(audio_file)

    return audio, cover


def get_reply_markup(track_name: str, artist: str) -> types.ReplyMarkupInlineKeyboard:
    """Generate a reply markup with the track name and update button."""
    return types.ReplyMarkupInlineKeyboard(
        [
            [
                types.InlineKeyboardButton(
                    text=track_name,
                    type=types.InlineKeyboardButtonTypeSwitchInline(
                        query=artist,
                        target_chat=types.TargetChatCurrent()
                    )
                )
            ],
            [
                types.InlineKeyboardButton(
                    text="Update ",
                    type=types.InlineKeyboardButtonTypeUrl("https://t.me/FallenProjects"),
                )
            ]
        ]
    )
