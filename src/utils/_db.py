import asyncio
import os
from typing import Optional

from pymongo import AsyncMongoClient
from pytdbot import types


from src.config import MONGO_URI, LOGGER_ID
from ._dataclass import Spotify


async def convert_to_m4a(input_file: str, cover_file: str, track: Spotify) -> str | None:
    """Convert audio to M4A with cover art and metadata."""
    abs_input = os.path.abspath(input_file)
    abs_cover = os.path.abspath(cover_file)
    output_file = f"{os.path.splitext(abs_input)[0]}.m4a"

    cmd = [
        "ffmpeg", "-y",
        "-i", abs_input, "-i", abs_cover,
        "-map", "0:a", "-map", "1:v",
        "-c:a", "aac", "-b:a", "192k",
        "-c:v", "png",
        "-metadata:s:v", "title=Album cover",
        "-metadata:s:v", "comment=Cover (front)",
        "-metadata", f"lyrics={track.lyrics}",
        "-metadata", f"title={track.name}",
        "-metadata", f"artist={track.artist}",
        "-metadata", f"album={track.album}",
        "-metadata", f"year={track.year}",
        "-metadata", "genre=Spotify",
        "-metadata", "comment=Via NoiNoi_bot | FallenProjects",
        "-f", "mp4",
        output_file,
    ]

    proc = await asyncio.create_subprocess_exec(
        *cmd, stdout=asyncio.subprocess.PIPE, stderr=asyncio.subprocess.PIPE
    )
    stdout, stderr = await proc.communicate()

    if proc.returncode != 0:
        print(f"ffmpeg failed:\n{stderr.decode(errors='ignore')}")
        print(f"stdout:\n{stdout.decode(errors='ignore')}")
        return None

    return output_file



class MongoDB:
    def __init__(self):
        self.mongo_client = AsyncMongoClient(MONGO_URI)
        self._db = self.mongo_client["SpTube"]
        self.songs = self._db["songs"]
        self.logger_chat_id = LOGGER_ID
        self._cache: dict[str, str] = {}

    async def connect(self) -> None:
        """Establish connection to MongoDB and load cache."""
        await self.mongo_client.aconnect()
        try:
            await self.mongo_client.admin.command("ping")
        except Exception as e:
            raise e
        await self._load_cache()

    async def _load_cache(self) -> None:
        """Load all stored songs into in-memory cache."""
        async for song in self.songs.find():
            self._cache[song["_id"]] = song["link"]

    async def store_song_link(self, track_id: str, link: str) -> None:
        """Store or update a song link in MongoDB and cache."""
        await self.songs.update_one({"_id": track_id}, {"$set": {"link": link}}, upsert=True)
        self._cache[track_id] = link

    async def get_song_link(self, track_id: str) -> Optional[str]:
        """Retrieve song link from cache or MongoDB."""
        if track_id in self._cache:
            return self._cache[track_id]

        song = await self.songs.find_one({"_id": track_id})
        if song:
            self._cache[track_id] = song["link"]
            return song["link"]
        return None

    async def get_song_file_id(self, track_id: str) -> Optional[str]:
        """Retrieve the Telegram file ID for a stored song link."""
        from src import client

        link = await self.get_song_link(track_id)
        if not link:
            return None

        info = await client.getMessageLinkInfo(url=link)
        if isinstance(info, types.Error) or not info.message:
            client.logger.warning(f"❌ Failed to get message link info: {getattr(info, 'message', info)}")
            return None

        msg = await client.getMessage(info.chat_id, info.message.id)
        if isinstance(msg, types.Error):
            client.logger.warning(f"❌ Failed to get message: {msg.message}")
            return None

        content = msg.content
        if isinstance(content, types.MessageAudio):
            return content.audio.audio.remote.id
        elif isinstance(content, types.MessageDocument):
            return content.document.document.remote.id
        elif isinstance(content, types.MessageVideo):
            return content.video.video.remote.id

        client.logger.warning(f"❌ Unsupported media type in stored link: {content}")
        await self.remove_song(track_id)
        return None

    async def upload_song_and_get_file_id(
            self, file_path: str, cover: Optional[str], track: Spotify
    ) -> Optional[str] | types.Error:
        """Upload song to logger chat, store link, and return file ID."""
        from src import client

        thumb = types.InputThumbnail(thumbnail=types.InputFileLocal(cover) if cover else types.InputFileRemote(track.cover), width=640, height=640)
        async def _send(path: str):
            return await client.sendAudio(
                chat_id=self.logger_chat_id,
                audio=types.InputFileLocal(path),
                album_cover_thumbnail=thumb,
                title=track.name,
                performer=track.artist,
                duration=track.duration,
                caption=f"<b>{track.name}</b>\n<i>{track.artist}</i>",
            )

        upload = await _send(file_path)

        # Handle "uploaded as voice" issue (TG MOOD)
        if not isinstance(upload, types.Error) and isinstance(upload.content, types.MessageVoiceNote):
            fixed_path = await convert_to_m4a(file_path, cover, track)
            if not fixed_path:
                public_link = await client.getMessageLink(upload.chat_id, upload.id)
                return types.Error(
                    message=f"Failed to upload audio - here is your song: {public_link.link or 'No Link 2x sed moment'}"
                )

            await upload.delete()
            upload = await _send(fixed_path)

            try:
                os.remove(fixed_path)
            except Exception as e:
                client.logger.warning(f"❌ Failed to remove converted file: {e}")

        if isinstance(upload, types.Error):
            client.logger.warning(f"❌ Failed to upload audio: {upload.message}")
            return upload

        public_link = await client.getMessageLink(upload.chat_id, upload.id)
        if isinstance(public_link, types.Error):
            client.logger.warning(f"❌ Failed to get public link: {public_link.message}")
            return public_link

        try:
            os.remove(file_path)
        except Exception as e:
            client.logger.warning(f"❌ Failed to remove original file: {e}")

        if isinstance(upload.content, types.MessageAudio):
            await self.store_song_link(track.tc, public_link.link)
            return upload.content.audio.audio.remote.id

        client.logger.info(f"file_path: {file_path} | cover: {cover}")
        client.logger.warning(f"❌ Unsupported media type in uploaded audio: {upload}")
        return types.Error(
            message=f"Failed to upload audio - here is your song: {public_link.link}"
        )

    async def remove_song(self, track_id: str) -> None:
        """Remove song from MongoDB and cache."""
        await self.songs.delete_one({"_id": track_id})
        if track_id in self._cache:
            del self._cache[track_id]

    async def close(self) -> None:
        """Close MongoDB connection and clear cache."""
        await self.mongo_client.aclose()
        self._cache.clear()


# Global DB instance
db: MongoDB = MongoDB()
