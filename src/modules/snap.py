from typing import Union, List, Optional
import asyncio

from pytdbot import Client, types
from src.utils import ApiData, Filter, APIResponse, Download
from ._utils import has_audio_stream


def batch_chunks(items: List[str], size: int = 10) -> List[List[str]]:
    return [items[i:i + size] for i in range(0, len(items), size)]


async def _handle_media_upload(
    client: Client,
    message: types.Message,
    media_url: str,
    media_type: str,
    reply_message: types.Message
) -> Optional[types.Error]:
    input_file = types.InputFileRemote(media_url)
    send_func = {
        "photo": message.reply_photo,
        "video": message.reply_video,
        "animation": message.reply_animation
    }.get(media_type)

    if not send_func:
        return types.Error(message="Unsupported media type")

    result = await send_func(**{media_type: input_file})
    if isinstance(result, types.Error) and "WEBPAGE_CURL_FAILED" in result.message:
        local_file = await Download(None).download_file(media_url, "")
        if isinstance(local_file, types.Error):
            return local_file
        input_file_local = types.InputFileLocal(local_file)
        result = await send_func(**{media_type: input_file_local})

    if isinstance(result, types.Error):
        client.logger.warning(f"❌ Media upload failed: {result.message}")
        return result

    return None


async def _send_media_album(
    client: Client,
    message: types.Message,
    media_urls: List[str],
    media_type: str
) -> Optional[types.Error]:
    content_cls = {
        "photo": types.InputMessagePhoto,
        "video": types.InputMessageVideo,
        "animation": types.InputMessageAnimation
    }.get(media_type)

    if not content_cls:
        return types.Error(message="Unsupported media type")

    contents = [
        content_cls(**{media_type: types.InputFileRemote(url)})
        for url in media_urls
    ]

    result = await client.sendMessageAlbum(
        chat_id=message.chat_id,
        input_message_contents=contents,
        reply_to=types.InputMessageReplyToMessage(message_id=message.id),
    )

    if isinstance(result, types.Error):
        client.logger.warning(f"❌ Media album upload failed: {result.message}")
        return result

    return None


@Client.on_message(filters=Filter.command("insta"))
async def insta_cmd(client: Client, message: types.Message) -> None:
    parts = message.text.split(" ", 1)
    if len(parts) < 2 or not parts[1].strip():
        await message.reply_text("Please provide a valid search query.")
        return None
    return await process_insta_query(client, message, parts[1].strip())


@Client.on_message(filters=Filter.save_snap())
async def insta_autodetect(client: Client, message: types.Message):
    return await process_insta_query(client, message, message.text.strip())


async def process_insta_query(client: Client, message: types.Message, query: str) -> None:
    reply = await message.reply_text("⏳ Processing...")
    api = ApiData(query)
    api_data: Union[APIResponse, types.Error, None] = await api.get_snap()

    if isinstance(api_data, types.Error):
        await reply.edit_text(f"❌ Error: {api_data.message}")
        return
    if not api_data:
        await reply.edit_text("❌ No results found.")
        return

    # --- Handle Images ---
    if api_data.image:
        for batch in batch_chunks(api_data.image, 10):
            error = await (
                _handle_media_upload(client, message, batch[0], "photo", reply)
                if len(batch) == 1
                else _send_media_album(client, message, batch, "photo")
            )
            if error:
                await reply.edit_text(f"❌ Failed to send photo(s): {error.message}")
                return

    # --- Handle Videos ---
    if api_data.video:
        video_urls = [v.video for v in api_data.video if v.video]
        if not video_urls:
            await reply.delete()
            return

        if len(video_urls) == 1:
            error = await _handle_media_upload(client, message, video_urls[0], "video", reply)
            if error:
                await reply.edit_text(f"❌ Failed to send video: {error.message}")
            else:
                await reply.delete()
            return

        # Check audio presence concurrently
        results = await asyncio.gather(
            *(has_audio_stream(url) for url in video_urls),
            return_exceptions=True
        )

        videos_with_audio, videos_without_audio = [], []
        for url, result in zip(video_urls, results):
            if isinstance(result, Exception):
                client.logger.warning(f"❌ Failed to check audio for {url}: {result}")
                continue
            (videos_with_audio if result else videos_without_audio).append(url)

        if not videos_with_audio and not videos_without_audio:
            await reply.edit_text("❌ No valid videos found.")
            return

        # Send videos with audio
        for batch in batch_chunks(videos_with_audio, 10):
            error = await (
                _handle_media_upload(client, message, batch[0], "video", reply)
                if len(batch) == 1
                else _send_media_album(client, message, batch, "video")
            )
            if error:
                await reply.edit_text(f"❌ Failed to send video(s): {error.message}")
                return

        if not videos_without_audio:
            await reply.delete()
            return

        # Send videos without audio as animations
        for batch in batch_chunks(videos_without_audio, 10):
            error = await (
                _handle_media_upload(client, message, batch[0], "animation", reply)
                if len(batch) == 1
                else _send_media_album(client, message, batch, "animation")
            )
            if error:
                await reply.edit_text(f"❌ Failed to send animation(s): {error.message}")
                return

    await reply.delete()
    return
