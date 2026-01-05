from pydantic import BaseModel
from typing import List, Optional, Union

# api/track?url=
class TrackResponse(BaseModel):
    id: str
    url: str
    cdnurl: str
    key: Optional[str] = None
    platform: str

class Track(BaseModel):
    title: str
    id: str
    url: str
    thumbnail: Optional[str] = None
    duration: int
    channel: Optional[str] = None
    views: Optional[str] = None
    platform: str

# api/get_url?url= or  /api/search
class SearchResponse(BaseModel):
    results: List[Track]

class SnapVideo(BaseModel):
    url: Optional[str] = None
    thumbnail: Optional[str] = None


class SnapAudio(BaseModel):
    url: Optional[str] = None

# api/snap?url=
class SnapResponse(BaseModel):
    videos: Optional[List[SnapVideo]] = None
    audios: Optional[List[SnapAudio]] = None
    images: Optional[List[str]] = None
    title: Optional[str] = None


# api/sp?url=
class Spotify(BaseModel):
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
