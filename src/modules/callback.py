from urllib.parse import urlparse

from pytdbot import Client, types

from src.utils import ApiData, shortener
from ._media_utils import process_track_media, get_reply_markup
from ._utils import handle_help_callback, StartMessage
from .start import get_main_menu_keyboard
from .. import db


@Client.on_updateNewCallbackQuery()
async def callback_query(c: Client, message: types.UpdateNewCallbackQuery):
    data = message.payload.data.decode()
    user_id = message.sender_user_id
    # Help menu
    if data.startswith("help_"):
        await handle_help_callback(c, message)
        return

    # Back to main menu
    if data == "back_menu":
        await message.answer("⏳ Returning to main menu…")
        bot_username = c.me.usernames.editable_username
        bot_name = c.me.first_name
        await message.edit_message_text(
            text=StartMessage.format(bot_name=bot_name, bot_username=bot_username),
            disable_web_page_preview=True,
            reply_markup=get_main_menu_keyboard(bot_username)
        )
        return

    # Only handle spot_ callbacks
    if not data.startswith("spot_"):
        await message.answer("Unexpected callback data", show_alert=True)
        await c.deleteMessages(message.chat_id, [message.message_id])
        return

    split1, split2 = data.find("_"), data.rfind("_")
    if split1 == -1 or split2 == -1 or split1 == split2:
        await c.deleteMessages(message.chat_id, [message.message_id])
        return

    id_enc, uid = data[split1 + 1: split2], data[split2 + 1:]
    if uid not in ("0", str(user_id)):
        await message.answer("🚫 This button wasn't meant for you.", show_alert=True)
        return

    url = shortener.decode_url(id_enc)
    if not url:
        await message.answer("Callback Expired", show_alert=True)
        await c.deleteMessages(message.chat_id, [message.message_id])
        return

    if "spotify.com" in url:
        parsed = urlparse(url)
        parts = parsed.path.strip("/").split("/")
        if len(parts) >= 2 and parts[0] == "track":
            if file_id := await db.get_song_file_id(parts[1]):
                audio = types.InputFileRemote(file_id)
                reply = await c.editMessageMedia(chat_id=message.chat_id, message_id=message.message_id, input_message_content=types.InputMessageAudio(audio=audio))
                if isinstance(reply, types.Error):
                    c.logger.error(f"Failed to send audio file: {reply.message}")
                    await message.edit_message_text(f"Failed to send the song. Please try again later.\n{reply.message}")
                return

    await message.answer("⏳ Processing your track, please wait...", show_alert=True)
    api = ApiData(url)
    track = await api.get_track()
    if isinstance(track, types.Error):
        await message.edit_message_text(f"Failed to fetch track: {track.message or 'Unknown error'}")
        return

    msg = await message.edit_message_text("🔄 Downloading the song...")
    if isinstance(msg, types.Error):
        c.logger.warning(f"❌ Failed to edit message: {msg.message}")
        return

    # Process the track media
    result = await process_track_media(c, track, chat_id=message.chat_id, message_id=message.message_id)
    if isinstance(result, types.Error):
        await message.edit_message_text(result.message)
        return

    audio, cover = result
    if not audio:
        await message.edit_message_text("No Audio")
        return

    # Send the audio
    reply = await c.editMessageMedia(
        chat_id=message.chat_id,
        message_id=message.message_id,
        input_message_content=types.InputMessageAudio(
            audio=audio,
            album_cover_thumbnail=types.InputThumbnail(types.InputFileLocal(cover)) if cover else None,
        ),
    )

    if isinstance(reply, types.Error):
        c.logger.error(f"Failed to send audio file: {reply.message}")
        await msg.edit_text(f"Failed to send the song. Please try again later.\n{reply.message}")
