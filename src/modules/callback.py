import re
from pytdbot import Client, types

from src.utils import ApiData, Download, shortener


@Client.on_updateNewCallbackQuery()
async def callback_query(c: Client, message: types.UpdateNewCallbackQuery):
    data = message.payload.data.decode()
    if data.startswith("help_"):
        await handle_help_callback(c, message)
        return

    elif data == "back_menu":
        get_msg = await message.getMessage()
        if isinstance(get_msg, types.Error):
            c.logger.warning(f"❌ Failed to get message: {get_msg.message}")
            return

        from .start import welcome
        await welcome(c, get_msg)
        await c.deleteMessages(message.chat_id, [message.message_id], revoke=True)
        return

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

async def handle_help_callback(_: Client, message: types.UpdateNewCallbackQuery):
    data = message.payload.data.decode()
    platform = data.replace("help_", "")

    examples = {
        "spotify": (
            "💡<b>Spotify Downloader</b>\n\n"
            "Songs available in high-quality 320kbps:\n\n"
            "👉 <code>https://open.spotify.com/track/3n3Ppam7vgaVa1iaRUc9Lp</code>\n"
            "👉 <code>https://open.spotify.com/artist/4YRxDV8wJFPHPTeXepOstw</code>\n"
            "👉 <code>https://open.spotify.com/album/2yYfIOq25JQWvUQ9AR172D</code>"
        ),
        "youtube": (
            "💡<b>YouTube Downloader</b>\n\n"
            "Download videos or audio:\n\n"
            "👉 <code>https://youtu.be/dQw4w9WgXcQ</code>\n"
            "👉 <code>https://www.youtube.com/watch?v=dQw4w9WgXcQ</code>"
        ),
        "soundcloud": (
            "💡<b>SoundCloud Downloader</b>\n\n"
            "Get high-quality music:\n\n"
            "👉 <code>https://soundcloud.com/yagihmael/mood-lofi-1?utm_source=clipboard&utm_medium=text&utm_campaign=social_sharing</code>\n"
            "👉 <code>https://soundcloud.com/yagihmael/mood-lofi-1</code>"
        ),
        "apple": (
            "💡<b>Apple Music Downloader</b>\n\n"
            "Lossless music links supported:\n\n"
            "👉 <code>https://music.apple.com/us/song/mood-lofi/1674035500</code>\n"
            "👉 <code>https://music.apple.com/us/album/ram-siya-ram-from-adipurush-hindi-single/1690001719</code>\n"
            "👉 <code>https://music.apple.com/us/album/glory/1763807339</code>\n"
            "👉 <code>https://music.apple.com/us/artist/yagihmael/1674035500</code>"
        ),
        "instagram": (
            "💡<b>Instagram Downloader</b>\n\n"
            "Download Reels, Stories, and Posts:\n\n"
            "👉 <code>https://www.instagram.com/reel/Cxyz123/</code>"
        ),
        "pinterest": (
            "💡<b>Pinterest Downloader</b>\n\n"
            "Photos and videos are available to download:\n\n"
            "👉 <code>https://www.pinterest.com/pin/1085649053904273177/</code>"
        ),
        "facebook": (
            "💡<b>Facebook Downloader</b>\n\n"
            "Works with videos from public pages:\n\n"
            "👉 <code>https://www.facebook.com/watch/?v=123456789</code>"
        ),
        "twitter": (
            "💡<b>Twitter Downloader</b>\n\n"
            "Download videos or Photos from posts:\n\n"
            "👉 <code>https://x.com/i/status/1951310276814578086</code>\n"
            "👉 <code>https://twitter.com/i/status/1951310276814578086</code>\n"
            "👉 <code>https://x.com/luismbat/status/1951307858764607604/photo/1</code>"
        ),
        "tiktok": (
            "💡<b>TikTok Downloader</b>\n\n"
            "Supports watermark-free download:\n\n"
            "👉 <code>https://vt.tiktok.com/ZSB3BovQp/</code>\n"
            "👉 <code>https://vt.tiktok.com/ZSSe7NprD/</code>"
        ),
        "threads": (
            "💡<b>Threads Downloader</b>\n\n"
            "Download media from Threads:\n\n"
            "👉 <code>https://www.threads.com/@camycavero/post/DM0FquaM2At?xmt=AQF0u_6ebeMHEjWCw0cm0Li4i8fI3INIU7YeSMffM9DmDw</code>"
        ),
    }

    reply_text = examples.get(platform, "<b>No help available for this platform.</b>")
    await message.answer(text="Help Menu")
    await message.edit_message_text(
        text=reply_text,
        parse_mode="html",
        disable_web_page_preview=True,
        reply_markup=types.ReplyMarkupInlineKeyboard([
            [
                types.InlineKeyboardButton(
                    text="⬅️ Back",
                    type=types.InlineKeyboardButtonTypeCallback("back_menu".encode())
                )
            ]
        ])
    )
