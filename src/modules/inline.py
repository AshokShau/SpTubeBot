from pytdbot import Client, types

from src import config
from src.utils import ApiData, Download


@Client.on_updateNewInlineQuery()
async def inline_search(c: Client, message: types.UpdateNewInlineQuery):
    query = message.query.strip()
    if not query:
        return

    api = ApiData(query)
    search = await api.search(limit="15")

    # Handle API error
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
        return

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
                id=track.id,
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


@Client.on_updateNewChosenInlineResult()
async def inline_result(c: Client, message: types.UpdateNewChosenInlineResult):
    result_id = message.result_id
    inline_message_id = message.inline_message_id

    # Can't edit if no inline_message_id is present
    if not inline_message_id:
        c.logger.warning(message)
        return

    # Fetch track data
    api = ApiData(result_id)
    track = await api.get_track()
    if isinstance(track, types.Error):
        return

    status_text = f"<b>🎵 {track.name}</b>\n👤 {track.artist} | 📀 {track.album}\n⏱️ {track.duration}s"
    parsed_status = await c.parseTextEntities(status_text, types.TextParseModeHTML())
    if isinstance(parsed_status, types.Error):
        c.logger.warning(f"❌ Text parse error: {parsed_status.message}")
        return

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
        return

    audio_file, cover = result
    caption = f"<b>{track.name}</b>\n<i>{track.artist}</i>"
    parsed_caption = await c.parseTextEntities(caption, types.TextParseModeHTML())

    upload = await c.sendAudio(
        chat_id=config.LOGGER_ID,
        audio=types.InputFileLocal(audio_file),
        # album_cover_thumbnail=types.InputThumbnail(types.InputFileLocal(cover)) if cover else None,
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
        return

    file_id = upload.content.audio.audio.remote.id
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
