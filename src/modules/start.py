import time
from datetime import datetime

from pytdbot import Client, types

from src import __version__, StartTime
from src.utils import Filter

@Client.on_message(filters=Filter.command(["start", "help"]))
async def start_cmd(c: Client, message: types.Message):
    text = (
        "<b>🎧 Welcome to SpTube Bot</b>\n\n"
        "This bot lets you <b>search</b>, <b>download</b>, and <b>stream</b> music from platforms like:\n"
        "• Spotify\n"
        "• YouTube\n"
        "• SoundCloud\n"
        "• Apple Music\n\n"
        "📥 Try sending a song name, playlist link, or use inline mode:\n"
        "<code>@SpTubeBot your search here</code>\n\n"
        "🔗 Source Code: <a href=\"https://github.com/AshokShau/SpTubeBot\">GitHub Repo</a>\n"
        "📜 /privacy — View Privacy Policy"
    )

    reply = await message.reply_text(text, parse_mode="html", disable_web_page_preview=True)
    if isinstance(reply, types.Error):
        c.logger.warning(f"Error sending start/help message: {reply.message}")


@Client.on_message(filters=Filter.command("privacy"))
async def privacy_handler(_: Client, message: types.Message):
    await message.reply_text(
        "🔒 <b>Privacy Policy</b>\n\n"
        "This bot does <b>not store</b> any personal data or chat history.\n"
        "All queries are processed in real time and nothing is logged.\n\n"
        "🛠️ <b>Open Source</b> — You can inspect and contribute:\n"
        "<a href=\"https://github.com/AshokShau/SpTubeBot\">github.com/AshokShau/SpTubeBot</a>",
        parse_mode="html",
        disable_web_page_preview=True
    )



@Client.on_message(filters=Filter.command("ping"))
async def ping_cmd(client: Client, message: types.Message) -> None:
    start_time = time.monotonic()
    reply_msg = await message.reply_text("🏓 Pinging...")
    latency = (time.monotonic() - start_time) * 1000  # in ms
    uptime = datetime.now() - StartTime
    uptime_str = str(uptime).split(".")[0]

    response = (
        "📊 <b>System Performance Metrics</b>\n\n"
        f"⏱️ <b>Bot Latency:</b> <code>{latency:.2f} ms</code>\n"
        f"🕒 <b>Uptime:</b> <code>{uptime_str}</code>\n"
        f"🤖 <b>Bot Version:</b> <code>{__version__}</code>"
    )
    done = await reply_msg.edit_text(response, disable_web_page_preview=True)
    if isinstance(done, types.Error):
        client.logger.warning(f"Error sending message: {done}")
    return None

