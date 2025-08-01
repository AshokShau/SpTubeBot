import time
from datetime import datetime

from pytdbot import Client, types

from src import __version__, StartTime
from src.utils import Filter

@Client.on_message(filters=Filter.command(["start", "help"]))
async def start_cmd(c: Client, message: types.Message):
    bot_username = c.me.usernames.editable_username
    text = (
        "<b>🎧 Welcome to SpTube Bot</b>\n\n"
        "Stream, download, and enjoy music from your favorite platforms:\n"
        "• <b>Spotify</b>\n"
        "• <b>YouTube</b>\n"
        "• <b>SoundCloud</b>\n"
        "• <b>Apple Music</b>\n\n"
        "<b>🎥 Now also supports media from:</b>\n"
        "• <b>Instagram</b> (Reels, Posts, Stories)\n"
        "• <b>Pinterest</b>\n"
        "• <b>Facebook</b> (Videos)\n"
        "• <b>TikTok</b>\n\n"
        f"<b>📚 Version:</b> <code>{__version__}</code>\n"
        "📥 <b>How to use:</b>\n"
        "• Send a song name, link, or media URL directly.\n"
        f"• Use inline mode: <code>@{bot_username} your search</code>\n\n"
        "📜 <b>Privacy Policy:</b> /privacy"
    )

    reply = await message.reply_text(
        text,
        parse_mode="html",
        disable_web_page_preview=True,
        reply_markup=types.ReplyMarkupInlineKeyboard(
            [
                [
                    types.InlineKeyboardButton(
                        text="Add me to your group",
                        type=types.InlineKeyboardButtonTypeUrl(
                            f"https://t.me/{bot_username}?startgroup=true"
                        ),
                    ),
                    types.InlineKeyboardButton(
                        text="GitHub",
                        type=types.InlineKeyboardButtonTypeUrl(
                            "https://github.com/AshokShau/SpTubeBot"
                        ),
                    )
                ]
            ]
        ),
    )
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

