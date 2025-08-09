import re
import uuid
from typing import Union

from pytdbot import Client, types

from src import config
from src.utils import ApiData, Download, upload_cache, shortener, APIResponse


@Client.on_updateNewInlineQuery()
async def inline_search(c: Client, message: types.UpdateNewInlineQuery):
    query = message.query.strip()
    if not query:
        return None
    api = ApiData(query)
    if api.is_save_snap_url():
        return await process_snap_inline(c, message, query)

    search = await api.get_info() if api.is_valid() else await api.search(limit="15")
    if isinstance(search, types.Error):
        await c.answerInlineQuery(
            message.id,
            results=[
                types.InputInlineQueryResultArticle(
                    id="error",
                    title="❌ Search Failed",
                    description=search.message or "Could not search Spotify.",
                )
            ]
        )
        return None

    results = []
    for track in search.results:
        display_text = (
            f"<b>🎧 Track:</b> <b>{track.name}</b>\n"
            f"<b>👤 Artist:</b> <i>{track.artist}</i>\n"
            f"<b>📅 Year:</b> {track.year}\n"
            f"<b>⏱ Duration:</b> {track.duration // 60}:{track.duration % 60:02d} mins\n"
            f"<b>🔗 Platform:</b> {track.platform.capitalize()}\n"
            f"<code>{track.id}</code>"
        )

        parse = await c.parseTextEntities(display_text, types.TextParseModeHTML())
        if isinstance(parse, types.Error):
            c.logger.warning(f"❌ Error parsing inline result for {track.name}: {parse.message}")
            continue

        reply_markup = types.ReplyMarkupInlineKeyboard(
            [
                [
                    types.InlineKeyboardButton(
                        text=f"{track.name}",
                        type=types.InlineKeyboardButtonTypeSwitchInline(query=track.artist, target_chat=types.TargetChatCurrent())
                    ),
                ],
            ]
        )

        results.append(
            types.InputInlineQueryResultArticle(
                id=shortener.encode_url(track.url),
                title=f"{track.name} - {track.artist}",
                description=f"{track.name} by {track.artist} ({track.year})",
                thumbnail_url=track.cover_small,
                input_message_content=types.InputMessageText(parse),
                reply_markup=reply_markup,
            )
        )

    response = await c.answerInlineQuery(
        message.id,
        results=results,
    )

    if isinstance(response, types.Error):
        c.logger.warning(f"❌ Inline response error: {response.message}")
    return None


@Client.on_updateNewChosenInlineResult()
async def inline_result(c: Client, message: types.UpdateNewChosenInlineResult):
    result_id = message.result_id
    inline_message_id = message.inline_message_id
    if not inline_message_id:
        return None

    # Fetch track data
    url = shortener.decode_url(result_id)
    if not url:
        return None

    api = ApiData(url)
    if api.is_save_snap_url():
        return None

    track = await api.get_track()
    if isinstance(track, types.Error):
        return None

    status_text = f"<b>🎵 {track.name}</b>\n👤 {track.artist} | 📀 {track.album}\n⏱️ {track.duration}s"
    parsed_status = await c.parseTextEntities(status_text, types.TextParseModeHTML())
    if isinstance(parsed_status, types.Error):
        c.logger.warning(f"❌ Text parse error: {parsed_status.message}")
        return None

    await c.editInlineMessageText(
        inline_message_id=inline_message_id,
        input_message_content=types.InputMessageText(parsed_status),
    )

    dl = Download(track)
    result = await dl.process()
    if isinstance(result, types.Error):
        error_text = await c.parseTextEntities(result.message, types.TextParseModeHTML())
        await c.editInlineMessageText(
            inline_message_id=inline_message_id,
            input_message_content=types.InputMessageText(error_text),
        )
        return None

    audio_file, cover = result
    caption = f"<b>{track.name}</b>\n<i>{track.artist}</i>"
    parsed_caption = await c.parseTextEntities(caption, types.TextParseModeHTML())
    cached_file_id = upload_cache.get(track.tc)
    if cached_file_id:
        await c.editInlineMessageMedia(
            inline_message_id=inline_message_id,
            input_message_content=types.InputMessageAudio(
                audio=types.InputFileRemote(cached_file_id),
                album_cover_thumbnail=types.InputThumbnail(types.InputFileLocal(cover)) if cover else None,
                title=track.name,
                performer=track.artist,
                duration=track.duration,
                caption=parsed_caption,
            ),
        )
        return None

    upload = await c.sendAudio(
        chat_id=config.LOGGER_ID,
        audio=types.InputFileLocal(audio_file),
        album_cover_thumbnail=types.InputThumbnail(types.InputFileLocal(cover)) if cover else None,
        title=track.name,
        performer=track.artist,
        duration=track.duration,
        caption=caption,
    )

    if isinstance(upload, types.Error):
        fallback_text = await c.parseTextEntities(upload.message, types.TextParseModeHTML())
        await c.editInlineMessageText(
            inline_message_id=inline_message_id,
            input_message_content=types.InputMessageText(fallback_text),
        )
        return None

    file_id = upload.content.audio.audio.remote.id
    upload_cache.set(track.tc, file_id)
    send_audio = await c.editInlineMessageMedia(
        inline_message_id=inline_message_id,
        input_message_content=types.InputMessageAudio(
            audio=types.InputFileRemote(file_id),
            album_cover_thumbnail=types.InputThumbnail(types.InputFileLocal(cover)) if cover else None,
            title=track.name,
            performer=track.artist,
            duration=track.duration,
            caption=parsed_caption,
        ),
    )

    if isinstance(send_audio, types.Error):
        c.logger.error(f"❌ Failed to send audio: {send_audio.message}")
        fallback_text = await c.parseTextEntities(send_audio.message, types.TextParseModeHTML())
        await c.editInlineMessageText(
            inline_message_id=inline_message_id,
            input_message_content=types.InputMessageText(fallback_text),
        )
        return None
    return None


def get_query_id():
    return str(uuid.uuid4())

async def process_snap_inline(c: Client, message: types.UpdateNewInlineQuery, query: str):
    api = ApiData(query)
    api_data: Union[APIResponse, types.Error, None] = await api.get_snap()

    if isinstance(api_data, types.Error) or not api_data:
        text = api_data.message.strip() or "An unknown error occurred."
        parse = await c.parseTextEntities(text, types.TextParseModeHTML())
        await c.answerInlineQuery(
            inline_query_id=message.id,
            results=[
                types.InputInlineQueryResultArticle(
                    id=get_query_id(),
                    title="❌ Search Failed",
                    description="Something went wrong.",
                    input_message_content=types.InputMessageText(text=parse)
                )
            ],
            cache_time=5
        )
        return

    results = []
    reply_markup = types.ReplyMarkupInlineKeyboard(
        [
            [
                types.InlineKeyboardButton(
                    text=f"Search Again",
                    type=types.InlineKeyboardButtonTypeSwitchInline(query=query, target_chat=types.TargetChatCurrent())
                ),
            ],
        ]
    )

    for idx, image_url in enumerate(api_data.image or []):
        if not image_url or not re.match("^https?://", image_url):
            continue

        results.append(
            types.InputInlineQueryResultPhoto(
                id=get_query_id(),
                photo_url=image_url,
                thumbnail_url=image_url,
                title=f"Photo {idx + 1}",
                description=f"Image result #{idx + 1}",
                input_message_content = types.InputMessagePhoto(photo=types.InputFileRemote(image_url)),
                reply_markup=reply_markup
            )
        )

    for idx, video_data in enumerate(api_data.video or []):
        video_url = getattr(video_data, 'video', None)
        thumb_url = getattr(video_data, 'thumbnail', '')
        if not video_url or not re.match("^https?://", video_url):
            continue

        results.append(
            types.InputInlineQueryResultVideo(
                id=get_query_id(),
                video_url=video_url,
                mime_type="video/mp4",
                thumbnail_url=thumb_url if thumb_url and re.match("^https?://", thumb_url) else "https://i.pinimg.com/736x/e2/c6/eb/e2c6eb0b48fc00f1304431bfbcacf50e.jpg",
                title=f"Video {idx + 1}",
                description=f"Video result #{idx + 1}",
                input_message_content=types.InputMessageVideo(
                    video=types.InputFileRemote(video_url),
                    thumbnail=types.InputThumbnail(types.InputFileRemote(thumb_url or video_url))
                ),
                reply_markup=reply_markup
            )
        )

    if len(results) == 0:
        parse = await c.parseTextEntities("No media found for this query", types.TextParseModeHTML())
        results.append(
            types.InputInlineQueryResultArticle(
                id=get_query_id(),
                title="No media found",
                description="Try a different search term",
                input_message_content=types.InputMessageText(text=parse)
            )
        )

    done = await c.answerInlineQuery(
        inline_query_id=message.id,
        results=results,
        cache_time=5,
    )
    if isinstance(done, types.Error):
        c.logger.error(f"❌ Failed to answer inline query: {done.message}")
        await c.answerInlineQuery(
            inline_query_id=message.id,
            results=[
                types.InputInlineQueryResultArticle(
                    id=get_query_id(),
                    title="❌ Search Failed",
                    description="Maybe Video size is too big.",
                    input_message_content=types.InputMessageText(text=done.message)
                )
            ],
            cache_time=5
        )
