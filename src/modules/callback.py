import re
from pytdbot import Client, types

from src.utils import ApiData, Download, shortener


@Client.on_updateNewCallbackQuery()
async def callback_query(c: Client, message: types.UpdateNewCallbackQuery):
    data = message.payload.data.decode()
    user_id = message.sender_user_id

    if not data.startswith("spot_"):
        c.logger.warning(f"⚠️ Invalid callback data received: {data}")
        return

    split1, split2 = data.find("_"), data.rfind("_")
    if split1 == -1 or split2 == -1 or split1 == split2:
        await message.answer("❌ Invalid callback format.", show_alert=True)
        return

    id_enc = data[split1 + 1: split2]
    uid = data[split2 + 1:]

    if uid != "0" and uid != str(user_id):
        await message.answer("🚫 This button wasn't meant for you.", show_alert=True)
        return

    url = shortener.decode_url(id_enc)
    if not url:
        await message.answer("⚠️ This button has expired. Please try again.", show_alert=True)
        return

    api = ApiData(url)
    track = await api.get_track()
    if isinstance(track, types.Error):
        await message.answer(f"❌ Failed to fetch track info.\n<b>{track.message}</b>", show_alert=True)
        return

    await message.answer("⏳ Processing your track, please wait...", show_alert=True)

    msg = await message.edit_message_text("🔄 Downloading the song...")
    if isinstance(msg, types.Error):
        c.logger.warning(f"❌ Failed to edit message: {msg.message}")
        return

    dl = Download(track)
    result = await dl.process()
    if isinstance(result, types.Error):
        await msg.edit_text(f"❌ Download failed.\n<b>{result.message}</b>")
        return

    audio_file, cover = result
    # Handle t.me links (media already uploaded in a channel)
    match = re.match(r"https?://t\.me/([^/]+)/(\d+)", audio_file)
    if match:
        info = await c.getMessageLinkInfo(audio_file)
        if isinstance(info, types.Error) or info.message is None:
            c.logger.error(f"❌ Failed to resolve link: {audio_file}")
            return

        dl_msg = await c.getMessage(info.chat_id, info.message.id)
        if isinstance(dl_msg, types.Error):
            c.logger.error(f"❌ Failed to fetch message: {dl_msg.message}")
            return

        file = await dl_msg.download()
        if isinstance(file, types.Error):
            c.logger.error(f"❌ Failed to download message ID {info.message.id}: {file.message}")
            return

        audio_file = file.path

    status_text = f"<b>🎵 {track.name}</b>\n👤 {track.artist} | 📀 {track.album}\n⏱️ {track.duration}s"
    parse = await c.parseTextEntities(status_text, types.TextParseModeHTML())
    reply_markup = types.ReplyMarkupInlineKeyboard(
        [
            [
                types.InlineKeyboardButton(
                    text=f"{track.name}",
                    type=types.InlineKeyboardButtonTypeUrl("https://t.me/FallenProjects"),
                ),
            ],
        ]
    )
    reply = await c.editMessageMedia(
        chat_id=message.chat_id,
        message_id=message.message_id,
        input_message_content=types.InputMessageAudio(
            audio=types.InputFileLocal(audio_file),
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
