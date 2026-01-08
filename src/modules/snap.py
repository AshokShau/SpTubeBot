import html
import uuid
from typing import Union, List, Optional
import asyncio

from pytdbot import Client, types
from pytdbot.exception import StopHandlers

from src.utils import ApiData, Filter, SnapResponse, Download

from ._fsub import fsub
from ._utils import has_audio_stream

def batch_chunks(items: List[str], size: int = 10) -> List[List[str]]:
    return [items[i:i + size] for i in range(0, len(items), size)]

async def _handle_media_upload(
    client: Client,
    message: types.Message,
    media_url: str,
    media_type: str,
    reply_message: types.Message,
    text: Optional[str] = None,
) -> Optional[types.Error]:
    send_func = {
        "photo": message.reply_photo,
        "video": message.reply_video,
        "animation": message.reply_animation,
        "audio": message.reply_audio,
    }.get(media_type)

    if not send_func:
        return types.Error(message="Unsupported media type")

    input_file = types.InputFileRemote(media_url)
    kwargs = {media_type: input_file, "caption": text}
    result = await send_func(**kwargs)
    if isinstance(result, types.Error) and "WEBPAGE_CURL_FAILED" in result.message:
        file_ext = ".mp4" if media_type in {"video", "animation"} else ".jpg"
        file_name = f"{uuid.uuid4()}{file_ext}"

        local_file = await Download(None).download_file(media_url, file_name)
        if isinstance(local_file, types.Error):
            client.logger.warning(f"❌ Media download failed: {local_file.message}")
            return local_file

        kwargs = {media_type: types.InputFileLocal(local_file), "caption": text}
        result = await send_func(**kwargs)
    if isinstance(result, types.Error):
        client.logger.warning(f"❌ Media upload failed: {result.message}")
        return result

    return None

async def _send_media_album(
    client: Client,
    message: types.Message,
    media_urls: List[str],
    media_type: str,
    caption: str = None,
) -> Optional[types.Error]:
    parsed_caption = await client.parseTextEntities(caption, types.TextParseModeHTML())
    if isinstance(parsed_caption, types.Error):
        client.logger.warning(f"Failed to parse caption: {parsed_caption.message}")
        parsed_caption = None

    input_contents = []
    for idx, url in enumerate(media_urls):
        input_file = types.InputFileRemote(url)
        caption_to_use = parsed_caption if idx == 0 else None
        if media_type == "photo":
            input_contents.append(
                types.InputMessagePhoto(
                    photo=input_file,
                    caption=caption_to_use,
                )
            )
        elif media_type == "video":
            input_contents.append(
                types.InputMessageVideo(
                    video=input_file,
                    caption=caption_to_use,
                )
            )
        elif media_type == "animation":
            input_contents.append(
                types.InputMessageAnimation(
                    animation=input_file,
                    caption=caption_to_use,
                )
            )
        elif media_type == "audio":
            input_contents.append(
                types.InputMessageAudio(
                    audio=input_file,
                    caption=caption_to_use,
                )
            )
        else:
            return types.Error(message="Unsupported media type")

    result = await client.sendMessageAlbum(
        chat_id=message.chat_id,
        input_message_contents=input_contents,
        reply_to=types.InputMessageReplyToMessage(message_id=message.id),
    )

    if isinstance(result, types.Error):
        client.logger.warning(
            f"❌ Media album upload failed: {result.message} \n"
            f"type: {media_type}"
        )
        return result

    return None

async def process_insta_query(client: Client, message: types.Message, query: str) -> None:
    reply = await message.reply_text("⏳ Processing...")
    api = ApiData(query)
    api_data: Union[SnapResponse, types.Error, None] = await api.get_snap()
    if isinstance(api_data, types.Error) or not api_data:
        await reply.edit_text(f"❌ Error: {api_data.message if api_data else 'No results found'}")
        return

    caption = html.escape(api_data.title) or "#FA"
    # --- Handle Images ---
    if api_data.images:
        for batch in batch_chunks(api_data.images, 10):
            error = await (
                _handle_media_upload(client, message, api_data.images[0], "photo", reply, caption)
                if len(api_data.images) == 1
                else _send_media_album(client, message, batch, "photo", caption)
            )

            if error:
                await reply.edit_text(f"❌ Failed to send photo(s): {error.message}")
                return

    # --- Handle Audio ---
    if api_data.audios:
        audio_urls = [a.url for a in api_data.audios if a.url]
        for batch in batch_chunks(audio_urls, 10):
            if error := await (
                    _handle_media_upload(client, message, batch[0], "audio", reply, caption)
                    if len(batch) == 1
                    else _send_media_album(client, message, batch, "audio", caption)
            ):
                client.logger.warning(f"❌ Failed to send audio(s): {error.message}")

    # --- Handle Videos ---
    if api_data.videos:
        video_urls = [v.url for v in api_data.videos if v.url]
        if not video_urls:
            await reply.delete()
            return

        if len(video_urls) == 1:
            if error := await _handle_media_upload(client, message, video_urls[0], "video", reply, caption):
                await reply.edit_text(f"❌ Failed to send video: {error.message}")
            else:
                await reply.delete()
            return

        results = await asyncio.gather(*(has_audio_stream(url) for url in video_urls), return_exceptions=True)
        videos_with_audio, videos_without_audio = [], []
        for url, result in zip(video_urls, results):
            if isinstance(result, Exception):
                client.logger.warning(f"❌ Failed to check audio for {url}: {result}")
                continue
            (videos_with_audio if result else videos_without_audio).append(url)

        if not videos_with_audio and not videos_without_audio:
            await reply.edit_text("❌ No valid videos found.")
            return

        for batch in batch_chunks(videos_with_audio, 10):
            error = await (
                _handle_media_upload(client, message, videos_with_audio[0], "video", reply, caption)
                if len(videos_with_audio) == 1
                else _send_media_album(client, message, batch, "video", caption)
            )
            if error:
                await reply.edit_text(f"❌ Failed to send video(s): {error.message}")
                return
            await asyncio.sleep(1)

        if not videos_without_audio:
            await reply.delete()
            return

        # Send videos without audio as animations
        for i, url in enumerate(videos_without_audio):
            current_caption = caption if i == 0 else None
            if error := await _handle_media_upload(client, message, url, "animation", reply, current_caption):
                client.logger.warning(f"❌ Failed to send animation: {error.message}")

    await reply.delete()
    return

@Client.on_message(filters=Filter.command("insta"))
@fsub
async def insta_cmd(client: Client, message: types.Message) -> None:
    parts = message.text.split(" ", 1)
    if len(parts) < 2 or not parts[1].strip():
        await message.reply_text("Please provide a valid search query.")
        return None

    api = ApiData(parts[1].strip())
    valid_url = api.extract_save_snap_url()
    if not valid_url:
        await message.reply_text("Please provide a valid search query.")
        return None

    await process_insta_query(client, message, valid_url)
    raise StopHandlers

@Client.on_message(filters=Filter.save_snap())
@fsub
async def insta_autodetect(client: Client, message: types.Message):
    api = ApiData(message.text.strip())
    valid_url = api.extract_save_snap_url()
    if not valid_url:
        return None
    await process_insta_query(client, message, valid_url)
    raise StopHandlers
