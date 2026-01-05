import asyncio
import os
import re

from pytdbot import Client, types

from src.config import DOWNLOAD_PATH
from src.utils import Filter
from ._fsub import fsub


@Client.on_message(filters=Filter.command(["yt", "youtube"]))
@fsub
async def youtube_cmd(c: Client, message: types.Message):
    parts = message.text.split(" ", 1)
    if len(parts) < 2:
        await message.reply_text("Please provide a search query.")
        return

    query = parts[1]
    if not re.match(r"^https?://", query):
        await message.reply_text("Please provide a valid URL.")
        return

    reply = await message.reply_text("🔍 Preparing download...")
    output_template = str(DOWNLOAD_PATH / "%(title).80s.%(ext)s")

    format_selector = "bestvideo[ext=mp4]+bestaudio[ext=m4a]/bestvideo+bestaudio/best"
    ytdlp_params = [
        "yt-dlp",
        "--no-warnings",
        "--no-playlist",
        "--quiet",
        "--geo-bypass",
        "--retries", "2",
        "--continue",
        "--no-part",
        "--restrict-filenames",
        "--concurrent-fragments", "3",
        "--socket-timeout", "10",
        "--throttled-rate", "100K",
        "--retry-sleep", "1",
        "--no-write-info-json",
        "--no-embed-metadata",
        "--no-embed-chapters",
        "--no-embed-subs",
        "--merge-output-format", "mp4",
        "--extractor-args", "youtube:player_js_version=actual",
        "-o", output_template,
        "-f", format_selector,
        "--print", "after_move:filepath",
        query
    ]

    proc = await asyncio.create_subprocess_exec(
        *ytdlp_params,
        stdout=asyncio.subprocess.PIPE,
        stderr=asyncio.subprocess.PIPE,
    )

    try:
        stdout, stderr = await asyncio.wait_for(proc.communicate(), timeout=300)
    except asyncio.TimeoutError:
        await reply.edit_text("⏳ Download timed out.")
        return

    if proc.returncode != 0:
        error_msg = stderr.decode().strip()
        if "is not a valid URL" in error_msg:
            await reply.edit_text("❌ Invalid URL provided. Please provide a valid video URL.")
        else:
            await reply.edit_text(f"❌ Error downloading:\n<code>{error_msg}</code>")
        return

    downloaded_path = stdout.decode().strip()
    if not downloaded_path:
        await reply.edit_text("❌ Could not find downloaded file.")
        return

    try:
        if not os.path.exists(downloaded_path):
            await reply.edit_text("❌ Downloaded file not found.")
            return

        done = await message.reply_video(
            video=types.InputFileLocal(downloaded_path),
            supports_streaming=True,
            caption="This video automatically deletes in 2 minutes so save or forward it now.",
        )
        if isinstance(done, types.Error):
            await reply.edit_text(f"❌ Error: {done.message}")
        else:
            await reply.delete()
            async def delete_message():
                await done.delete()
            c.loop.call_later(120, lambda: asyncio.create_task(delete_message()))
    finally:
        try:
            os.remove(downloaded_path)
        except Exception as e:
            c.logger.error(f"Error deleting file {downloaded_path}: {e}")
