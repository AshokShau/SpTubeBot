import os
from typing import Optional
from pymongo import AsyncMongoClient
from pytdbot import types
from src.config import MONGO_URI, LOGGER_ID
from ._dataclass import TrackInfo

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
        return None

    async def upload_song_and_get_file_id(
        self, file_path: str, cover: Optional[str], track: TrackInfo
    ) -> Optional[str]:
        """Upload song to logger chat, store link, and return file ID."""
        from src import client

        upload = await client.sendAudio(
            chat_id=self.logger_chat_id,
            audio=types.InputFileLocal(file_path),
            album_cover_thumbnail=types.InputThumbnail(types.InputFileLocal(cover)) if cover else None,
            title=track.name,
            performer=track.artist,
            duration=track.duration,
            caption=f"<b>{track.name}</b>\n<i>{track.artist}</i>",
        )
        if isinstance(upload, types.Error):
            client.logger.warning(f"❌ Failed to upload audio: {upload.message}")
            return None

        public_link = await client.getMessageLink(upload.chat_id, upload.id)
        if isinstance(public_link, types.Error):
            client.logger.warning(f"❌ Failed to get public link: {public_link.message}")
            return None

        await self.store_song_link(track.tc, public_link.link)
        try:
            os.remove(file_path)
        except Exception as e:
            client.logger.warning(f"❌ Failed to remove file: {e}")
        return upload.content.audio.audio.remote.id

    async def close(self) -> None:
        """Close MongoDB connection and clear cache."""
        await self.mongo_client.aclose()
        self._cache.clear()


# Global DB instance
db: MongoDB = MongoDB()
