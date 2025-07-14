from pytdbot import Client, types

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

        results.append(
            types.InputInlineQueryResultArticle(
                id=track.id,
                title=f"{track.name} - {track.artist}",
                description=f"{track.name} by {track.artist} ({track.year})",
                thumbnail_url=track.cover_small,
                input_message_content=types.InputMessageText(parse),
            )
        )

    response = await c.answerInlineQuery(
        message.id,
        results=results,
        cache_time=5,
    )

    if isinstance(response, types.Error):
        c.logger.warning(f"❌ Inline response error: {response.message}")



@Client.on_updateNewChosenInlineResult()
async def inline_result(c: Client, message: types.UpdateNewChosenInlineResult):
    print(message)
    result_id = message.result_id
    inline_message_id = message.inline_message_id
    api = ApiData(result_id)
    track = await api.get_track()
    if isinstance(track, types.Error):
        return

    status_text = f"⏳ Downloading <b>{track.name}</b> by <i>{track.artist}</i>..."
    parse = await c.parseTextEntities(status_text, types.TextParseModeHTML())
    if isinstance(parse, types.Error):
        c.logger.warning(f"Text parse error: {parse.message}")
        return

    update_msg = await c.editInlineMessageText(
        inline_message_id=inline_message_id,
        input_message_content=types.InputMessageText(parse),
    )

    if isinstance(update_msg, types.Error):
        c.logger.warning(f"Failed to update message: {update_msg.message}")
        return

    # Download audio
    downloader = Download(track)
    audio_file, cover = await downloader.process()

    if not audio_file:
        return

    send_audio = await c.editInlineMessageMedia(
        inline_message_id=inline_message_id,
        input_message_content=types.InputMessageAudio(
            audio=types.InputFileLocal(audio_file),
            album_cover_thumbnail=types.InputThumbnail(types.InputFileLocal(cover)) if cover else None,
            title=track.name,
            performer=track.artist,
            duration=track.duration,
        ),
    )

    if isinstance(send_audio, types.Error):
        c.logger.error(f"❌ Failed to send audio: {send_audio.message}")
