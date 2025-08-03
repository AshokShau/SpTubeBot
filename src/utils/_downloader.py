import asyncio
import binascii
import logging
import re
import shutil

import time
import uuid
import zipfile
from pathlib import Path
from urllib.parse import urlparse as parse_url
from typing import Optional, Tuple, Union

from Crypto.Cipher import AES
from pytdbot import types

from src import config

from ._api import ApiData, _executor, HttpClient
from ._dataclass import TrackInfo, PlatformTracks, MusicTrack

# Constants
MAX_COVER_SIZE = 10 * 1024 * 1024  # 10MB
CHUNK_SIZE = 1024 * 1024 * 2  # 2MB
DEFAULT_FILE_PERM = 0o644

# Configure logging
logger = logging.getLogger(__name__)


class MissingKeyError(Exception):
    pass


class InvalidHexKeyError(Exception):
    pass


class Download:
    def __init__(self, track: Optional[TrackInfo]):
        self.track = track
        self.downloads_dir = Path(config.DOWNLOAD_PATH)
        self.downloads_dir.mkdir(parents=True, exist_ok=True, mode=0o755)

    async def process(self) -> Union[Tuple[str, Optional[str]], types.Error]:
        """Process the track download with optimized flow."""
        try:
            if not self.track.cdnurl:
                return types.Error(message="Missing CDN URL")

            # Check for existing files first
            output_file = self.downloads_dir / f"{self.track.tc}.ogg"
            cover_path = self.downloads_dir / f"{self.track.tc}_cover.jpg"

            if output_file.exists():
                logger.debug(f"Using cached file: {output_file}")
                return str(output_file), str(cover_path) if cover_path.exists() else None

            # Process based on platform
            if self.track.platform in ["youtube", "soundcloud"]:
                return await self.process_direct_dl()

            return await self.process_standard()

        except Exception as e:
            logger.error(f"Error processing track {self.track.tc}: {str(e)}", exc_info=True)
            return types.Error(message=f"Track processing failed: {str(e)}")

    async def process_direct_dl(self) -> Tuple[str, Optional[str]]:
        """Handle direct downloads with optimized flow."""
        if re.match(r'^https:\/\/t\.me\/([a-zA-Z0-9_]{5,})\/(\d+)$', self.track.cdnurl):
            cover_path = await self.save_cover(self.track.cover)
            return self.track.cdnurl, cover_path

        file_path = await self.download_file(self.track.cdnurl, "")
        cover_path = await self.save_cover(self.track.cover)
        return file_path, cover_path

    async def process_standard(self) -> Tuple[str, Optional[str]]:
        """Optimized standard processing flow."""
        start_time = time.monotonic()
        output_file = self.downloads_dir / f"{self.track.tc}.ogg"

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
                output_file.chmod(DEFAULT_FILE_PERM)
                return result
            finally:
                encrypted_file.unlink(missing_ok=True)
                decrypted_file.unlink(missing_ok=True)
        finally:
            logger.info(f"Processed {self.track.tc} in {time.monotonic() - start_time:.2f}s")

    async def download_and_decrypt(self, encrypted_path: Path, decrypted_path: Path) -> None:
        """Optimized download and decrypt pipeline using httpx."""
        client = await HttpClient.get_client()

        try:
            async with client.stream('GET', self.track.cdnurl) as response:
                if response.status_code != 200:
                    raise Exception(f"Unexpected status code: {response.status_code}")

                with encrypted_path.open('wb') as f:
                    async for chunk in response.aiter_bytes(CHUNK_SIZE):
                        f.write(chunk)

            decrypted_data = await asyncio.get_event_loop().run_in_executor(
                _executor,
                self._decrypt_file,
                encrypted_path,
                self.track.key
            )
            decrypted_path.write_bytes(decrypted_data)
        except Exception as e:
            encrypted_path.unlink(missing_ok=True)
            decrypted_path.unlink(missing_ok=True)
            raise e

    @staticmethod
    def _decrypt_file(file_path: Path, hex_key: str) -> bytes:
        try:
            key = binascii.unhexlify(hex_key)
            audio_aes_iv = binascii.unhexlify("72e067fbddcbcf77ebe8bc643f630d93")
            cipher = AES.new(key, AES.MODE_CTR, nonce=b'', initial_value=audio_aes_iv)

            with file_path.open('rb') as f:
                return cipher.decrypt(f.read())
        except binascii.Error as e:
            raise InvalidHexKeyError(f"Invalid hex key: {e}")

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
        output_file = self.downloads_dir / f"{self.track.tc}.ogg"

        try:
            await self._run_ffmpeg(input_file, output_file)
            await self._add_vorbis_comments(output_file)
        except Exception as e:
            output_file.unlink(missing_ok=True)
            raise e

        return str(output_file), cover_path

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
        except FileNotFoundError:
            raise Exception("ffmpeg not found in PATH")

    async def _add_vorbis_comments(self, output_file: Path) -> None:
        """Optimized vorbis comment addition."""
        if not shutil.which('vorbiscomment'):
            logger.warning("vorbiscomment not found - skipping metadata")
            return

        metadata = [
            f"ALBUM={self.track.album}",
            f"ARTIST={self.track.artist}",
            f"TITLE={self.track.name}",
            f"GENRE=Spotify @FallenProjects",
            f"YEAR={self.track.year}",
            f"TRACKNUMBER={self.track.tc}",
            f"COMMENT=By @FallenProjects",
            f"PUBLISHER={self.track.artist}",
            f"DURATION={self.track.duration}",
        ]
        metadata.insert(0, f"METADATA_BLOCK_PICTURE={await self._create_vorbis_image_block()}")
        tmp_file = self.downloads_dir / f"{uuid.uuid4()}.txt"
        try:
            tmp_file.write_text("\n".join(metadata))

            cmd = ['vorbiscomment', '-a', str(output_file), '-c', str(tmp_file)]
            proc = await asyncio.create_subprocess_exec(*cmd,
                                                        stdout=asyncio.subprocess.PIPE,
                                                        stderr=asyncio.subprocess.PIPE,
                                                        )
            _, stderr = await proc.communicate()

            if proc.returncode != 0:
                raise Exception(f"vorbiscomment failed: {stderr.decode()}")
        finally:
            tmp_file.unlink(missing_ok=True)

    async def _create_vorbis_image_block(self) -> str:
        path = await self.save_cover(self.track.cover)
        if not path:
            return ""

        image_path = Path(path)
        base64_path = image_path.with_suffix(".base64")

        try:
            process = await asyncio.create_subprocess_exec(
                "./cover_gen.sh", str(image_path),
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE,
            )
            stdout, stderr = await process.communicate()

            if process.returncode != 0:
                logging.error(f"cover_gen.sh failed with code {process.returncode}\nOutput: {stderr.decode().strip()}")
                return ""

            if not base64_path.exists():
                logging.error(f"{base64_path} not generated by cover_gen.sh")
                return ""

            return base64_path.read_text()

        except Exception as e:
            logging.error(f"Error generating vorbis block: {e}")
            return ""
        finally:
            base64_path.unlink(missing_ok=True)

    async def download_file(self, url: str, file_path: str = "") -> str | types.Error:
        if not url:
            return types.Error(code=400, message="No URL provided")

        file_path = Path(file_path) if file_path else self._generate_filename(url)

        if file_path.exists():
            return str(file_path)

        client = await HttpClient.get_client()

        try:
            async with client.stream("GET", url, follow_redirects=True) as response:
                if response.status_code != 200:
                    return types.Error(
                        code=response.status_code,
                        message=f"Unexpected status code: {response.status_code}"
                    )

                with file_path.open("wb") as f:
                    async for chunk in response.aiter_bytes(CHUNK_SIZE):
                        f.write(chunk)

            file_path.chmod(DEFAULT_FILE_PERM)
            return str(file_path)

        except Exception as e:
            file_path.unlink(missing_ok=True)
            return types.Error(code=500, message=f"Download failed: {str(e)}")

    def _generate_filename(self, url: str) -> Path:
        """Generate a safe filename from URL."""
        try:
            filename = Path(parse_url(url).path).name
            if not filename or filename == '/':
                filename = f"{uuid.uuid4()}.tmp"
            return self.downloads_dir / self._sanitize_filename(filename)
        except Exception:
            return self.downloads_dir / f"{uuid.uuid4()}.tmp"

    @staticmethod
    def _sanitize_filename(name: str) -> str:
        """Sanitize filename with minimal regex."""
        return re.sub(r'[^\w\-_. ]', '_', name)

    async def save_cover(self, cover_url: Optional[str]) -> Optional[str]:
        if not cover_url:
            return None

        cover_path = self.downloads_dir / f"{self.track.tc}_cover.jpg"
        if cover_path.exists():
            return str(cover_path)

        try:
            client = await HttpClient.get_client()
            response = await client.get(cover_url)
            if response.status_code != 200:
                return None

            cover_data = response.content
            if len(cover_data) > MAX_COVER_SIZE:
                logger.warning(f"Cover too large ({len(cover_data)} bytes)")
                return None

            cover_path.write_bytes(cover_data)
            cover_path.chmod(DEFAULT_FILE_PERM)
            return str(cover_path)
        except Exception as e:
            logger.warning(f"Failed to download cover: {e}")
            return None


async def download_playlist_zip(playlist: PlatformTracks) -> Optional[str]:
    """Optimized playlist download with parallel processing."""
    downloads_dir = Path(config.DOWNLOAD_PATH)
    zip_path = downloads_dir / f"playlist_{int(time.time())}.zip"
    temp_dir = downloads_dir / f"tmp_{uuid.uuid4()}"
    temp_dir.mkdir(parents=True, exist_ok=True)

    try:
        tasks = []
        for music in playlist.results:
            tasks.append(_process_playlist_track(music, temp_dir))

        audio_files = await asyncio.gather(*tasks, return_exceptions=True)
        audio_files = [f for f in audio_files if isinstance(f, Path)]

        if not audio_files:
            return None

        temp_zip = downloads_dir / f"tmp_{uuid.uuid4()}.zip"
        with zipfile.ZipFile(temp_zip, 'w', zipfile.ZIP_DEFLATED) as zipf:
            for file in audio_files:
                zipf.write(file, arcname=file.name)

        temp_zip.rename(zip_path)
        return str(zip_path)
    finally:
        for file in temp_dir.glob('*'):
            file.unlink(missing_ok=True)


async def _process_playlist_track(music: MusicTrack, temp_dir: Path) -> Optional[Path]:
    """Process a single playlist track."""
    try:
        track = await ApiData(music.url).get_track()
        if isinstance(track, types.Error):
            return None

        dl = Download(track)
        result = await dl.process()
        if isinstance(result, types.Error):
            return None

        audio_file, _ = result
        src_path = Path(audio_file)
        dest_path = temp_dir / src_path.name
        src_path.rename(dest_path)
        return dest_path
    except Exception as e:
        logger.warning(f"Failed to process track {music.url}: {e}")
        return None
