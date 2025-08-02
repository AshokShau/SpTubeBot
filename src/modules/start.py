import time
from datetime import datetime

from pytdbot import Client, types

from src import StartTime
from src.utils import Filter

def get_main_menu_keyboard(bot_username: str) -> types.ReplyMarkupInlineKeyboard:
    return types.ReplyMarkupInlineKeyboard([
        [
            types.InlineKeyboardButton(
                text="➕ Add to Group",
                type=types.InlineKeyboardButtonTypeUrl(
                    url=f"https://t.me/{bot_username}?startgroup=true"
                )
            ),
            types.InlineKeyboardButton(
                text="📂 GitHub",
                type=types.InlineKeyboardButtonTypeUrl(
                    url="https://github.com/AshokShau/SpTubeBot"
                )
            )
        ],
        [
            types.InlineKeyboardButton(
                text="Spotify",
                type=types.InlineKeyboardButtonTypeCallback("help_spotify".encode())
            ),
            types.InlineKeyboardButton(
                text="YouTube",
                type=types.InlineKeyboardButtonTypeCallback("help_youtube".encode())
            )
        ],
        [
            types.InlineKeyboardButton(
                text="SoundCloud",
                type=types.InlineKeyboardButtonTypeCallback("help_soundcloud".encode())
            ),
            types.InlineKeyboardButton(
                text="Apple Music",
                type=types.InlineKeyboardButtonTypeCallback("help_apple".encode())
            )
        ],
        [
            types.InlineKeyboardButton(
                text="Instagram",
                type=types.InlineKeyboardButtonTypeCallback("help_instagram".encode())
            ),
            types.InlineKeyboardButton(
                text="Pinterest",
                type=types.InlineKeyboardButtonTypeCallback("help_pinterest".encode())
            )
        ],
        [
            types.InlineKeyboardButton(
                text="Facebook",
                type=types.InlineKeyboardButtonTypeCallback("help_facebook".encode())
            ),
            types.InlineKeyboardButton(
                text="Twitter",
                type=types.InlineKeyboardButtonTypeCallback("help_twitter".encode())
            )
        ],
        [
            types.InlineKeyboardButton(
                text="TikTok",
                type=types.InlineKeyboardButtonTypeCallback("help_tiktok".encode())
            ),
            types.InlineKeyboardButton(
                text="Threads",
                type=types.InlineKeyboardButtonTypeCallback("help_threads".encode())
            )
        ],
    ])



@Client.on_message(filters=Filter.command("start"))
async def welcome(c: Client, message: types.Message):
    bot_username = c.me.usernames.editable_username
    text = (
        "<b>🎧 Welcome to SpTube Bot</b>\n"
        "Easily download music & media from your favorite platforms.\n\n"
        "📩 Send a name, link, or media URL\n"
        f"🔎 Try inline: <code>@{bot_username} your search</code>\n\n"
        "🔐 Privacy: /privacy"
    )
    reply = await message.reply_text(
        text,
        parse_mode="html",
        disable_web_page_preview=True,
        reply_markup=get_main_menu_keyboard(bot_username)
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
        f"⏱️ <b>Uptime:</b> <code>{uptime_str}</code>\n"
        f"👤 <b>Developer:</b> <a href=\"https://t.me/AshokShau\">@AshokShau</a>"
    )
    done = await reply_msg.edit_text(response, disable_web_page_preview=True)
    if isinstance(done, types.Error):
        client.logger.warning(f"Error sending message: {done}")
    return None

