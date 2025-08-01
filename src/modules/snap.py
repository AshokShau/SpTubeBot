from typing import Union, List
from pytdbot import Client, types
from src.utils import ApiData, Filter, APIResponse, Download

from ._utils import has_audio_stream

def batch_chunks(items: List[str], size: int = 10) -> List[List[str]]:
    return [items[i:i + size] for i in range(0, len(items), size)]


@Client.on_message(filters=Filter.command("insta"))
async def insta_cmd(client: Client, message: types.Message) -> None:
    parts = message.text.split(" ", 1)
    if len(parts) < 2:
        await message.reply_text("Please provide a search query.")
        return None

    query = parts[1]
    return await process_insta_query(client, message, query)

@Client.on_message(filters=Filter.save_snap())
async def insta_autodetect(c: Client, message: types.Message):
    return await process_insta_query(c, message, message.text)


async def process_insta_query(client: Client, message: types.Message, query: str) -> None:
    api = ApiData(query)
    api_data: Union[APIResponse, types.Error, None] = await api.get_snap()

    if isinstance(api_data, types.Error):
        await message.reply_text(f"Error: {api_data.message}")
        return

    if not api_data:
        client.logger.warning(f"No snap found for {query}")
        return

    # --- Image Handling ---
    images = api_data.image or []
    for batch in batch_chunks(images, 10):
        if len(batch) == 1:
            done = await message.reply_photo(photo=types.InputFileRemote(batch[0]))
            if isinstance(done, types.Error) and "WEBPAGE_CURL_FAILED" in done.message:
                dl = Download(None)
                local_file = await dl.download_file(batch[0], "")
                if isinstance(local_file, types.Error):
                    await message.reply_text(f"Image Error: {local_file.message}")
                    return

                await message.reply_photo(photo=types.InputFileLocal(local_file))
        else:
            done = await client.sendMessageAlbum(
                chat_id=message.chat_id,
                input_message_contents=[
                    types.InputMessagePhoto(photo=types.InputFileRemote(url))
                    for url in batch
                ],
                reply_to=types.InputMessageReplyToMessage(message_id=message.id),
            )
        if isinstance(done, types.Error):
            await message.reply_text(f"Image Error: {done.message}")
            return

    # --- Video Handling ---
    video_urls = [v.video for v in api_data.video if v.video]
    if not video_urls:
        return

    if len(video_urls) == 1:
        url = video_urls[0]
        done = await message.reply_video(video=types.InputFileRemote(url))
        if isinstance(done, types.Error) and "WEBPAGE_CURL_FAILED" in done.message:
            dl = Download(None)
            local_file = await dl.download_file(url, "")
            if isinstance(local_file, types.Error):
                await message.reply_text(f"Video Error: {local_file.message}")
                return

            await message.reply_video(video=types.InputFileLocal(local_file))
        return

    videos_with_audio = []
    videos_without_audio = []

    for url in video_urls:
        if await has_audio_stream(url):
            videos_with_audio.append(url)
        else:
            videos_without_audio.append(url)

    # Videos with audio
    for batch in batch_chunks(videos_with_audio, 10):
        if len(batch) == 1:
            done = await message.reply_video(video=types.InputFileRemote(batch[0]))
        else:
            done = await client.sendMessageAlbum(
                chat_id=message.chat_id,
                input_message_contents=[
                    types.InputMessageVideo(video=types.InputFileRemote(url))
                    for url in batch
                ],
                reply_to=types.InputMessageReplyToMessage(message_id=message.id),
            )
        if isinstance(done, types.Error):
            await message.reply_text(f"Video Error: {done.message}")
            return

    # Videos without audio
    for batch in batch_chunks(videos_without_audio, 10):
        if len(batch) == 1:
            done = await message.reply_animation(animation=types.InputFileRemote(batch[0]))
        else:
            done = await client.sendMessageAlbum(
                chat_id=message.chat_id,
                input_message_contents=[
                    types.InputMessageAnimation(animation=types.InputFileRemote(url))
                    for url in batch
                ],
                reply_to=types.InputMessageReplyToMessage(message_id=message.id),
            )
        if isinstance(done, types.Error):
            await message.reply_text(f"Animation Error: {done.message}")
            return
