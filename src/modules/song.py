from pytdbot import Client, types
from pytdbot.exception import StopHandlers

from src.utils import ApiData, shortener, Filter
from ._fsub import fsub


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
            text=f"{track.title} - {track.channel}",
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
@fsub
async def spotify_cmd(_: Client, message: types.Message):
    parts = message.text.split(" ", 1)
    if len(parts) < 2:
        await message.reply_text("Please provide a search query.")
        return

    query = parts[1]
    await process_spotify_query(message, query)
    raise StopHandlers


@Client.on_message(filters=Filter.sp_tube())
@fsub
async def spotify_autodetect(_: Client, message: types.Message):
    await process_spotify_query(message, message.text)
    raise StopHandlers

