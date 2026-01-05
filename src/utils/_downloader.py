import asyncio
import base64
import binascii
import logging
import mimetypes
import re
import time
import uuid
import zipfile
from pathlib import Path
from typing import Optional, Tuple, Union
from urllib.parse import urlparse

from Crypto.Cipher import AES
from mutagen.flac import Picture
from mutagen.oggvorbis import OggVorbis
from pytdbot import types

from src import config
from ._api import ApiData, HttpClient
from ._dataclass import Spotify, Track, TrackResponse

CHUNK_SIZE = 1024 * 1024  # 1 MB
DEFAULT_FILE_PERM = 0o644

# Configure logging
logger = logging.getLogger(__name__)


class MissingKeyError(Exception):
    pass


class InvalidHexKeyError(Exception):
    pass


class Download:
    def __init__(self, track: Optional[Union[Spotify, TrackResponse]] = None):
        self.track = track
        self.downloads_dir = config.DOWNLOAD_PATH
        self.downloads_dir.mkdir(parents=True, exist_ok=True, mode=0o755)
        if isinstance(track, TrackResponse):
            filename = str(uuid.uuid4())
        else:
            filename = str(uuid.uuid4()) if track is None else self._sanitize_filename(track.name)
        self.dl_file_name = filename
        self.output_file = self.downloads_dir / f"{filename}.ogg"

    async def process(self) -> Union[Tuple[str, Optional[str]], types.Error]:
        """Process the track download with optimized flow."""
        try:
            if not self.track.cdnurl:
                return types.Error(message="Missing CDN URL")

            if self.track.platform.lower() == "spotify":
                cover_path = self.downloads_dir / f"{self.track.tc}_cover.jpg"
                if self.output_file.exists():
                    logger.debug(f"Using cached file: {self.output_file}")
                    return str(self.output_file), str(cover_path) if cover_path.exists() else None
                return await self.process_standard()

            return await self.process_direct_dl()
        except Exception as e:
            logger.error(f"Error processing track {self.track.id}: {str(e)}", exc_info=True)
            return types.Error(message=f"Track processing failed: {str(e)}")

    async def process_direct_dl(self) -> Tuple[str, Optional[str]]:
        """Handle direct downloads with optimized flow."""
        if re.match(r'^https:\/\/t\.me\/([a-zA-Z0-9_]{5,})\/(\d+)$', self.track.cdnurl):
            return self.track.cdnurl, None

        file_path = await self.download_file(self.track.cdnurl, "")
        return file_path, None

    async def process_standard(self) -> Tuple[str, Optional[str]]:
        """Optimized standard processing flow."""
        start_time = time.monotonic()
        if not self.track.key:
            raise MissingKeyError("Missing CDN key")

        try:
            # Use temporary files in the same directory for atomic writes
            encrypted_file = self.downloads_dir / f"{uuid.uuid4()}.enc"
            decrypted_file = self.downloads_dir / f"{uuid.uuid4()}.tmp"

            try:
                await self.download_and_decrypt(encrypted_file, decrypted_file)
                await self.rebuild_ogg(decrypted_file)
                result = await self.vorb_repair_ogg(decrypted_file)
                self.output_file.chmod(DEFAULT_FILE_PERM)
                logger.info(f"Successfully processed {self.track.tc}: {result}")
                return result
            finally:
                encrypted_file.unlink(missing_ok=True)
                decrypted_file.unlink(missing_ok=True)
        finally:
            logger.info(f"Processed {self.track.tc} in {time.monotonic() - start_time:.2f}s")

    async def download_and_decrypt(self, encrypted_path: Path, decrypted_path: Path) -> None:
        client = await HttpClient.get_client()

        try:
            async with client.stream("GET", self.track.cdnurl) as response:
                response.raise_for_status()
                with encrypted_path.open("wb") as f:
                    async for chunk in response.aiter_bytes(CHUNK_SIZE):
                        f.write(chunk)

            decrypted_data = await asyncio.to_thread(
                self._decrypt_file, encrypted_path, self.track.key
            )
            decrypted_path.write_bytes(decrypted_data)

        except Exception:
            encrypted_path.unlink(missing_ok=True)
            decrypted_path.unlink(missing_ok=True)
            raise

    @staticmethod
    def _decrypt_file(file_path: Path, hex_key: str) -> bytes:
        try:
            key = binascii.unhexlify(hex_key)
            audio_aes_iv = binascii.unhexlify("72e067fbddcbcf77ebe8bc643f630d93")
            cipher = AES.new(key, AES.MODE_CTR, nonce=b'', initial_value=audio_aes_iv)

            with file_path.open('rb') as f:
                return cipher.decrypt(f.read())
        except binascii.Error as e:
            raise InvalidHexKeyError(f"Invalid hex key: {e}") from e

    @staticmethod
    async def rebuild_ogg(filename: Path) -> None:
        """Optimized OGG header rebuilding."""
        patches = {
            0: b'OggS',
            6: b'\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00',
            26: b'\x01\x1E\x01vorbis',
            39: b'\x02',
            40: b'\x44\xAC\x00\x00',
            48: b'\x00\xE2\x04\x00',
            56: b'\xB8\x01',
            58: b'OggS',
            62: b'\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00',
        }

        # Read-modify-write in one operation
        with filename.open('r+b') as file:
            for offset, data in patches.items():
                file.seek(offset)
                file.write(data)

    async def vorb_repair_ogg(self, input_file: Path) -> Tuple[str, Optional[str]]:
        """Optimized metadata adding with error handling."""
        cover_path = await self.save_cover(self.track.cover)
        try:
            await self._run_ffmpeg(input_file, self.output_file)
            await self._add_comments(self.output_file)
        except Exception as e:
            self.output_file.unlink(missing_ok=True)
            raise e

        return str(self.output_file), cover_path

    async def _run_ffmpeg(self, input_file: Path, output_file: Path) -> None:
        """Run ffmpeg with optimized parameters."""
        cmd = [
            'ffmpeg', '-y', '-hide_banner', '-loglevel', 'error',
            '-i', str(input_file),
            '-c', 'copy',
            '-metadata', f'lyrics={self.track.lyrics}',
            str(output_file)
        ]

        try:
            proc = await asyncio.create_subprocess_exec(*cmd,
                                                        stdout=asyncio.subprocess.PIPE,
                                                        stderr=asyncio.subprocess.PIPE,
                                                        )
            _, stderr = await proc.communicate()

            if proc.returncode != 0:
                raise Exception(f"ffmpeg failed: {stderr.decode()}")
        except FileNotFoundError as e:
            raise Exception("ffmpeg not found in PATH") from e

    async def _add_comments(self, output_file: Path) -> None:
        try:
            audio = OggVorbis(output_file)
            audio.clear()
            cover_path = await self.save_cover(self.track.cover)
            if cover_path:
                try:
                    image_data = Path(cover_path).read_bytes()
                    picture = Picture()
                    picture.type = 3  # Cover (front)
                    picture.mime = "image/jpeg"
                    picture.desc = "Cover (front)"
                    picture.width = 0
                    picture.height = 0
                    picture.depth = 0
                    picture.data = image_data
                    audio["METADATA_BLOCK_PICTURE"] = [base64.b64encode(picture.write()).decode("ascii")]
                except Exception as e:
                    logger.warning(f"Failed to embed cover: {e}")

            # --- Text Metadata ---
            audio["ALBUM"] = [self.track.album]
            audio["ARTIST"] = [self.track.artist]
            audio["TITLE"] = [self.track.name]
            audio["GENRE"] = ["Spotify"]
            audio["DATE"] = [str(self.track.year)]
            audio["ALBUMARTIST"] = [self.track.artist]
            audio["YEAR"] = [str(self.track.year)]
            audio["LYRICS"] = [self.track.lyrics]
            audio["TRACKNUMBER"] = [str(self.track.tc)]
            audio["COMMENT"] = ["Via NoiNoi_bot | FallenProjects"]
            audio["PUBLISHER"] = [self.track.artist]
            audio["DURATION"] = [str(self.track.duration)]
            audio.save()
        except Exception as e:
            logger.error(f"Failed to add comments: {e}", exc_info=True)

    async def download_file(self, url: str, file_name: str = "") -> str | types.Error:
        if not url:
            return types.Error(code=400, message="No URL provided")

        client = await HttpClient.get_client()
        self.downloads_dir.mkdir(parents=True, exist_ok=True)
        if not file_name:
            file_name = f"{self._sanitize_filename(self.dl_file_name)}.mp4"

        try:
            async with client.stream("GET", url, follow_redirects=True) as response:
                if response.status_code != 200:
                    return types.Error(
                        code=response.status_code,
                        message=f"Unexpected status code: {response.status_code}"
                    )

                cd = response.headers.get("content-disposition", "")
                if "filename=" in cd:
                    file_name = cd.split("filename=")[-1].strip('"')

                elif not file_name:
                    parsed_url = urlparse(url)
                    if url_name := Path(parsed_url.path).name:
                        file_name = url_name

                if not file_name or "." not in file_name:
                    ext = ""
                    if content_type := response.headers.get("content-type"):
                        if guessed_ext := mimetypes.guess_extension(
                            content_type.split(";")[0].strip()
                        ):
                            ext = guessed_ext
                    file_name = file_name or f"file-{uuid.uuid4().hex}{ext}"

                file_name = Path(file_name)
                file_path = self.downloads_dir / file_name
                temp_path = file_path.with_suffix(f"{file_path.suffix}.part")

                if file_path.exists():
                    return str(file_path)

                with temp_path.open("wb") as f:
                    async for chunk in response.aiter_bytes(CHUNK_SIZE):
                        f.write(chunk)

            temp_path.rename(file_path)
            file_path.chmod(DEFAULT_FILE_PERM)
            return str(file_path)

        except Exception as e:
            try:
                temp_path.unlink(missing_ok=True)
            except NameError:
                pass
            return types.Error(code=500, message=f"Download failed: {str(e)}")

    @staticmethod
    def _sanitize_filename(name: str) -> str:
        """Sanitize filename for cross-platform safety."""
        name = re.sub(r'[\x00-\x1f\x7f]+', '', name)
        name = re.sub(r'[<>:"/\\|?*]', '_', name)
        name = re.sub(r'[^\w\s\-.]', '_', name)
        name = re.sub(r'[\s_]+', ' ', name).strip()
        return name

    async def save_cover(self, cover_url: Optional[str]) -> Optional[str]:
        if not cover_url:
            logger.warning("No cover URL provided")
            return None

        cover_path = self.downloads_dir / f"{self.track.tc}_cover.jpg"
        if cover_path.exists():
            return str(cover_path)

        try:
            client = await HttpClient.get_client()
            response = await client.get(cover_url)
            if response.status_code != 200:
                logger.warning(f"Failed to download cover: HTTP {response.status_code}")
                return None

            cover_data = response.content
            cover_path.write_bytes(cover_data)
            cover_path.chmod(DEFAULT_FILE_PERM)
            return str(cover_path)
        except Exception as e:
            logger.warning(f"Failed to download cover: {e}")
            return None
