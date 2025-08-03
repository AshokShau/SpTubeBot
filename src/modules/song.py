import re

from pytdbot import Client, types

from src.utils import ApiData, shortener, Filter, download_playlist_zip


async def process_spotify_query(message: types.Message, query: str):
    response = await message.reply_text("⏳ Searching for tracks...")
    if isinstance(response, types.Error):
        await message.reply_text(f"Error: {response.message}")
        return

    api = ApiData(query)
    song_data = await api.get_info() if api.is_valid() else await api.search(limit="5")
    if isinstance(song_data, types.Error):
        await response.edit_text(f"❌ Error: {song_data.message}")
        return

    if not song_data or not song_data.results:
        await response.edit_text("❌ No results found.")
        return

    keyboard = [
        [types.InlineKeyboardButton(
            text=f"{track.name} - {track.artist}",
            type=types.InlineKeyboardButtonTypeCallback(
                f"spot_{shortener.encode_url(track.url)}_0".encode()
            )
        )]
        for track in song_data.results
    ]

    await response.edit_text(
        f"Search results for: <b>{query}</b>\n\nPlease tap on the song you want to download.",
        parse_mode="html",
        disable_web_page_preview=True,
        reply_markup=types.ReplyMarkupInlineKeyboard(keyboard),
    )


@Client.on_message(filters=Filter.command(["spot", "spotify", "song"]))
async def spotify_cmd(_: Client, message: types.Message):
    parts = message.text.split(" ", 1)
    if len(parts) < 2:
        await message.reply_text("Please provide a search query.")
        return

    query = parts[1]
    await process_spotify_query(message, query)


@Client.on_message(filters=Filter.sp_tube())
async def spotify_autodetect(_: Client, message: types.Message):
    await process_spotify_query(message, message.text)



@Client.on_message(filters=Filter.command(["dl_zip", "playlist"]))
async def dl_playlist(c: Client, message: types.Message):
    parts = message.text.strip().split(" ", 1)

    if len(parts) < 2 or not parts[1].strip():
        await message.reply_text("❗ Please provide a playlist URL or search query.\n\nExample: `/dl_zip artist or url`", parse_mode="markdown")
        return

    query = parts[1].strip()
    api = ApiData(query)
    result = await api.get_info()
    if isinstance(result, types.Error):
        await message.reply_text(f"❌ Failed to fetch playlist.\n<b>{result.message}</b>", parse_mode="html")
        return

    if not result.results:
        await message.reply_text("⚠️ No tracks found for the given input.")
        return

    first_song = result.results[0]
    if first_song.platform == "youtube":
        await message.reply_text("you cant dl yt playlist")
        return

    if len(result.results) > 15:
        await message.reply_text("⚠️ The playlist contains more than 15 tracks. Please download each track individually.")
        return

    reply = await message.reply_text(f"⏳ Downloading {len(result.results)} tracks and creating ZIP…")

    zip_path = await download_playlist_zip(result)
    if not zip_path:
        await message.reply_text("❌ Failed to download any tracks. Please try again.")
        return

    ok = await c.editMessageMedia(
        chat_id=reply.chat_id,
        message_id=reply.id,
        input_message_content=types.InputMessageDocument(types.InputFileLocal(zip_path)),
    )
    if isinstance(ok, types.Error):
        await message.reply_text(f"❌ Error: {ok.message}")
        return
