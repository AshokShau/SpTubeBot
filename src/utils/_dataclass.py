from pydantic import BaseModel
from typing import List


class TrackInfo(BaseModel):
    cdnurl: str
    key: str
    name: str
    artist: str
    tc: str
    cover: str
    lyrics: str
    album: str
    year: int
    duration: int
    platform: str

class MusicTrack(BaseModel):
    name: str
    artist: str
    id: str
    url: str
    year: str
    cover: str
    cover_small: str
    duration: int
    platform: str

class PlatformTracks(BaseModel):
    results: List[MusicTrack]
