import binascii
import logging
import re
import shutil
import subprocess
import time
import uuid
import zipfile
from pathlib import Path
from typing import Optional
from typing import Tuple, Union

import aiohttp
from Crypto.Cipher import AES
from pytdbot import types

from src import config
from src.utils._api import ApiData
from ._dataclass import TrackInfo, PlatformTracks

# Constants
DEFAULT_DOWNLOAD_DIR_PERM = 0o755
DEFAULT_FILE_PERM = 0o644
MAX_COVER_SIZE = 10 * 1024 * 1024  # 10MB
DOWNLOAD_TIMEOUT = 300  # 5 minutes in seconds

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


# Custom Exceptions
class MissingCDNURLError(Exception):
    pass


class MissingKeyError(Exception):
    pass


class InvalidHexKeyError(Exception):
    pass


class InvalidAESIVError(Exception):
    pass


class VorbisCommentNotFoundError(Exception):
    pass



class Download:
    def __init__(self, track: TrackInfo):
        self.track = track

    async def process(self) -> Union[Tuple[str, Optional[str]], types.Error]:
        try:
            if not self.track.cdnurl:
                return types.Error(message="Missing CDN URL")

            if self.track.platform in ["youtube", "soundcloud"]:
                return await self.process_direct_dl()

            return await self.process_standard()

        except Exception as e:
            logger.error(f"Error processing track {self.track.tc}: {e}")
            return types.Error(message=f"Track processing failed: {e}")

    async def process_direct_dl(self) -> Tuple[str, Optional[str]]:  # Changed return type
        if re.match(r'^https:\/\/t\.me\/([a-zA-Z0-9_]{5,})\/(\d+)$', self.track.cdnurl):
            cover_path = await self.save_cover(self.track.cover)
            return self.track.cdnurl, cover_path

        file_path = await self.download_file(self.track.cdnurl, "")
        cover_path = await self.save_cover(self.track.cover)
        return file_path, cover_path

    async def process_standard(self) -> Tuple[str, Optional[str]]:  # Changed return type
        downloads_dir = Path(config.DOWNLOAD_PATH)
        output_file = downloads_dir / f"{self.track.tc}.ogg"

        if output_file.exists():
            logger.info(f"✅ Found existing file: {output_file}")
            cover_path = downloads_dir / f"{self.track.tc}_cover.jpg"
            return str(output_file), str(cover_path) if cover_path.exists() else None

        if not self.track.key:
            raise MissingKeyError("Missing CDN key")

        start_time = time.time()

        try:
            encrypted_file = downloads_dir / f"{self.track.tc}.encrypted"
            decrypted_file = downloads_dir / f"{self.track.tc}_decrypted.ogg"

            try:
                await self.download_and_decrypt(encrypted_file, decrypted_file)
                await self.rebuild_ogg(decrypted_file)
                return await self.vorb_repair_ogg(decrypted_file)
            finally:
                encrypted_file.unlink(missing_ok=True)
                decrypted_file.unlink(missing_ok=True)
        finally:
            logger.info(f"Process completed in {time.time() - start_time:.2f}s")

    async def download_and_decrypt(self, encrypted_path: Path, decrypted_path: Path) -> None:
        # Download file
        async with aiohttp.ClientSession() as session:
            async with session.get(self.track.cdnurl) as resp:
                if resp.status != 200:
                    raise Exception(f"Unexpected status code: {resp.status}")

                data = await resp.read()

        # Write encrypted file
        encrypted_path.write_bytes(data)

        # Decrypt file
        start_time = time.time()
        decrypted_data = await self.decrypt_audio_file(encrypted_path, self.track.key)
        logger.info(f"Decryption completed in {(time.time() - start_time) * 1000:.2f}ms")

        # Write decrypted file
        decrypted_path.write_bytes(decrypted_data)

    async def get_cover(self) -> Optional[bytes]:
        cover_url = self.track.cover
        if not cover_url:
            return None

        timeout = aiohttp.ClientTimeout(total=30)
        headers = {
            "User-Agent": "Mozilla/5.0",
        }

        async with aiohttp.ClientSession(timeout=timeout, headers=headers) as session:
            async with session.get(cover_url) as resp:
                if resp.status != 200:
                    raise Exception(f"Unexpected status code: {resp.status}")
                cover_data = await resp.read()
                return cover_data

    async def decrypt_audio_file(self, file_path: Path, hex_key: str) -> bytes:
        if not file_path.exists():
            raise FileNotFoundError(f"File not found: {file_path}")

        try:
            key = binascii.unhexlify(hex_key)
        except binascii.Error as e:
            raise InvalidHexKeyError(f"Invalid hex key: {e}")

        data = file_path.read_bytes()

        try:
            audio_aes_iv = binascii.unhexlify("72e067fbddcbcf77ebe8bc643f630d93")
        except binascii.Error as e:
            raise InvalidAESIVError(f"Invalid AES IV: {e}")

        cipher = AES.new(key, AES.MODE_CTR, nonce=b'', initial_value=audio_aes_iv)
        decrypted = cipher.decrypt(data)
        return decrypted

    async def rebuild_ogg(self, filename: Path) -> None:
        with filename.open('r+b') as file:
            # OGG header structure patches
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

            for offset, data in patches.items():
                file.seek(offset)
                file.write(data)

    async def vorb_repair_ogg(self, input_file: Path) -> Tuple[str, Optional[str]]:  # Changed return type
        cover_path = await self.save_cover(self.track.cover)  # Changed to save_cover
        output_file = Path(config.DOWNLOAD_PATH) / f"{self.track.tc}.ogg"

        # Use ffmpeg to add metadata
        cmd = [
            'ffmpeg', '-i', str(input_file),
            '-c', 'copy',
            '-metadata', f'lyrics={self.track.lyrics}',
            str(output_file)
        ]

        try:
            subprocess.run(cmd, check=True, capture_output=True)
        except subprocess.CalledProcessError as e:
            raise Exception(f"ffmpeg failed: {e}\nOutput: {e.stderr.decode()}")

        try:
            await self.add_vorbis_comments(output_file, cover_path)
        except Exception as e:
            logger.error(f"Failed to add vorbis comments: {e}")

        return str(output_file), cover_path

    async def add_vorbis_comments(self, output_file: Path, cover_path: Optional[str]) -> None:
        """Modified to work with cover path instead of bytes"""
        if not shutil.which('vorbiscomment'):
            raise VorbisCommentNotFoundError("vorbiscomment not found")

        cover_data = None
        if cover_path:
            try:
                with open(cover_path, 'rb') as f:
                    cover_data = f.read()
            except Exception as e:
                logger.error(f"Failed to read cover file: {e}")

        metadata = (
            f"METADATA_BLOCK_PICTURE={await self.create_vorbis_image_block(cover_data)}\n"
            f"ALBUM={self.track.album}\n"
            f"ARTIST={self.track.artist}\n"
            f"TITLE={self.track.name}\n"
            f"GENRE=Spotify @FallenProjects\n"
            f"YEAR={self.track.year}\n"
            f"TRACKNUMBER={self.track.tc}\n"
            f"COMMENT=By @FallenProjects\n"
            f"PUBLISHER={self.track.artist}\n"
            f"DURATION={self.track.duration}\n"
        )

        tmp_file = Path("vorbis.txt")
        try:
            tmp_file.write_text(metadata)

            cmd = ['vorbiscomment', '-a', str(output_file), '-c', str(tmp_file)]
            subprocess.run(cmd, check=True, capture_output=True)
        finally:
            tmp_file.unlink(missing_ok=True)

    async def create_vorbis_image_block(self, image_bytes: Optional[bytes]) -> str:
        if not image_bytes:
            return ""

        tmp_cover = Path("cover.jpg")
        tmp_base64 = Path("cover.base64")

        try:
            tmp_cover.write_bytes(image_bytes)

            cmd = ["./cover_gen.sh", str(tmp_cover)]
            subprocess.run(cmd, check=True, capture_output=True)

            return tmp_base64.read_text()
        except Exception as e:
            logger.error(f"Failed to generate cover: {e}")
            return ""
        finally:
            tmp_cover.unlink(missing_ok=True)
            tmp_base64.unlink(missing_ok=True)

    async def download_file(self, url_str: str, file_path: str, overwrite: bool = False) -> str:
        if not url_str:
            raise ValueError("Empty URL provided")

        downloads_dir = Path(config.DOWNLOAD_PATH)

        # Determine filename
        if not file_path:
            file_path = self.determine_filename(url_str, None)  # Simplified

        file_path = Path(file_path)

        # Skip if file exists and not overwriting
        if not overwrite and file_path.exists():
            return str(file_path)

        # Ensure directory exists
        file_path.parent.mkdir(parents=True, exist_ok=True, mode=DEFAULT_DOWNLOAD_DIR_PERM)

        # Download to temp file first
        temp_path = file_path.with_suffix(file_path.suffix + '.part')

        try:
            async with aiohttp.ClientSession(timeout=aiohttp.ClientTimeout(total=DOWNLOAD_TIMEOUT)) as session:
                async with session.get(url_str) as resp:
                    if resp.status != 200:
                        raise Exception(f"Unexpected status code: {resp.status}")

                    with temp_path.open('wb') as f:
                        async for chunk in resp.content.iter_chunked(8192):
                            f.write(chunk)

            # Rename temp file to final name
            temp_path.rename(file_path)
            return str(file_path)
        except Exception:
            temp_path.unlink(missing_ok=True)
            raise

    def determine_filename(self, url_str: str, content_disp: Optional[str]) -> str:
        # Try from Content-Disposition first
        if filename := self.extract_filename(content_disp):
            return str(Path(config.DOWNLOAD_PATH) / self.sanitize_filename(filename))

        # Fall back to URL path
        try:
            from urllib.parse import urlparse
            parsed = urlparse(url_str)
            filename = Path(parsed.path).name

            if not filename or filename == '/' or '?' in filename:
                filename = f"{uuid.uuid4()}.tmp"

            return str(Path(config.DOWNLOAD_PATH) / self.sanitize_filename(filename))
        except Exception:
            return str(Path(config.DOWNLOAD_PATH) / f"{uuid.uuid4()}.tmp")

    def extract_filename(self, content_disp: Optional[str]) -> Optional[str]:
        if not content_disp:
            return None

        if match := re.search(r'(?i)filename\*?=[\'"]?(?:UTF-\d[\'"]*)?([^\'";\n]*)[\'"]?', content_disp):
            return match.group(1)
        return None

    def sanitize_filename(self, name: str) -> str:
        return re.sub(r'[/\\:*?"<>|]', '_', name)

    async def save_cover(self, cover_url: str) -> Optional[str]:
        """Downloads cover and saves to file, returns path"""
        if not cover_url:
            return None

        downloads_dir = Path(config.DOWNLOAD_PATH)
        cover_path = downloads_dir / f"{self.track.tc}_cover.jpg"
        if cover_path.exists():
            return str(cover_path)

        cover_data = await self.get_cover()
        if not cover_data:
            return None

        try:
            cover_path.write_bytes(cover_data)
            return str(cover_path)
        except Exception as e:
            logger.error(f"Failed to save cover: {e}")
            return None


async def download_playlist_zip(playlist: PlatformTracks) -> Optional[str]:
    downloads_dir = Path(config.DOWNLOAD_PATH)
    # TODO improve playlist name
    zip_path = downloads_dir / "playlist.zip"
    zip_temp_dir = downloads_dir / f"playlist_{int(time.time())}"
    zip_temp_dir.mkdir(parents=True, exist_ok=True)

    audio_files = []

    for music in playlist.results:
        # Get full track info using ApiData
        track = await ApiData(music.url).get_track()
        if isinstance(track, types.Error):
            continue  # Skip if failed

        # Download the track
        dl = Download(track)
        result = await dl.process()
        if isinstance(result, types.Error):
            continue  # Skip if failed

        audio_file, _ = result
        file_path = Path(audio_file)
        if file_path.exists():
            dest_path = zip_temp_dir / file_path.name
            file_path.rename(dest_path)
            audio_files.append(dest_path)

    # If nothing downloaded
    if not audio_files:
        return None

    # Create zip
    with zipfile.ZipFile(zip_path, 'w', zipfile.ZIP_DEFLATED) as zipf:
        for file in audio_files:
            zipf.write(file, arcname=file.name)

    # Cleanup
    for file in audio_files:
        # file.unlink(missing_ok=True)
        file.rmdir()

    zip_temp_dir.rmdir()
    return str(zip_path)
