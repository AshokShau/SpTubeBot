import re
from pytdbot import Client, types

from src.utils import ApiData, Download, shortener, db


from ._utils import handle_help_callback, StartMessage
from .start import get_main_menu_keyboard


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

    await message.answer("⏳ Processing your track, please wait...", show_alert=True)
    url = shortener.decode_url(id_enc)
    if not url:
        await c.deleteMessages(message.chat_id, [message.message_id])
        return

    # Get track info
    api = ApiData(url)
    track = await api.get_track()
    if isinstance(track, types.Error):
        await message.edit_message_text(f"❌ Failed to fetch track: {track.message or 'Unknown error'}")
        return

    msg = await message.edit_message_text("🔄 Downloading the song...")
    if isinstance(msg, types.Error):
        c.logger.warning(f"❌ Failed to edit message: {msg.message}")
        return

    reply_markup = types.ReplyMarkupInlineKeyboard(
        [
            [
                types.InlineKeyboardButton(
                    text=(
                        f'{track.name[:20]}...'
                        if len(track.name) > 20
                        else track.name
                    ),
                    type=types.InlineKeyboardButtonTypeUrl(
                        "https://t.me/FallenProjects"
                    ),
                )
            ]
        ]
    )
    status_text = f"<b>🎵 {track.name}</b>\n👤 {track.artist} | 📀 {track.album}\n⏱️ {track.duration}s"
    parse = await c.parseTextEntities(status_text, types.TextParseModeHTML())

    audio_file, cover, audio = None, None, None

    # Spotify shortcut if file already cached
    if track.platform.lower() == "spotify" and track.tc:
        if file_id := await db.get_song_file_id(track.tc):
            audio = types.InputFileRemote(file_id)

    # Download if not found in DB
    if not audio:
        dl = Download(track)
        result = await dl.process()
        if isinstance(result, types.Error):
            await msg.edit_text(f"❌ Download failed.\n<b>{result.message}</b>")
            return

        audio_file, cover = result
        if not audio_file:
            await msg.edit_text("❌ Failed to download song.\nPlease report this to @FallenProjects.")
            return

        if track.platform.lower() == "spotify":
            file_id = await db.upload_song_and_get_file_id(audio_file, cover, track)
            if not file_id:
                await msg.edit_text("❌ Failed to send song to database.")
                return

            audio = types.InputFileRemote(file_id)
        elif re.match(r"https?://t\.me/([^/]+)/(\d+)", audio_file):
            info = await c.getMessageLinkInfo(audio_file)
            if isinstance(info, types.Error) or not info.message:
                c.logger.error(f"❌ Failed to resolve link: {audio_file}")
                return

            public_msg = await c.getMessage(info.chat_id, info.message.id)
            if isinstance(public_msg, types.Error):
                c.logger.error(f"❌ Failed to fetch message: {public_msg.message}")
                await msg.edit_text(f"❌ Failed to fetch message: {public_msg.message}")
                return

            if isinstance(public_msg.content, types.MessageAudio):
                audio = types.InputFileRemote(public_msg.content.audio.audio.remote.id)
            elif isinstance(public_msg.content, types.MessageDocument):
                audio = types.InputFileRemote(public_msg.content.document.document.remote.id)
            elif isinstance(public_msg.content, types.MessageVideo):
                audio = types.InputFileRemote(public_msg.content.video.video.remote.id)
            else:
                c.logger.error(f"❌ No audio file in t.me link: {audio_file}")
                await msg.edit_text("⚠️ Audio file not found in t.me link")
                return
        else:
            audio = types.InputFileLocal(audio_file)

    reply = await c.editMessageMedia(
        chat_id=message.chat_id,
        message_id=message.message_id,
        input_message_content=types.InputMessageAudio(
            audio=audio,
            album_cover_thumbnail=types.InputThumbnail(types.InputFileLocal(cover)) if cover else None,
            title=track.name,
            performer=track.artist,
            duration=track.duration,
            caption=parse,
        ),
        reply_markup=reply_markup,
    )

    if isinstance(reply, types.Error):
        c.logger.error(f"❌ Failed to send audio file: {reply.message}")
        await msg.edit_text("❌ Failed to send the song. Please try again later.")
