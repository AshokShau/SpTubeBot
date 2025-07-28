# 🎵 SongBot - Telegram Music Downloader Bot  

A high-performance Telegram bot for downloading songs from Spotify, YouTube, and other platforms in premium quality.

<p align="center">
  <a href="https://github.com/AshokShau/SpTubeBot/stargazers">
    <img src="https://img.shields.io/github/stars/AshokShau/SpTubeBot?style=for-the-badge&logo=github&color=yellow" alt="Stars"/>
  </a>
  <a href="https://github.com/AshokShau/SpTubeBot/network/members">
    <img src="https://img.shields.io/github/forks/AshokShau/SpTubeBot?style=for-the-badge&logo=github" alt="Forks"/>
  </a>
  <a href="https://github.com/AshokShau/SpTubeBot/issues">
    <img src="https://img.shields.io/github/issues/AshokShau/SpTubeBot?style=for-the-badge&logo=github" alt="Issues"/>
  </a>
  <a href="https://github.com/AshokShau/SpTubeBot/releases">
    <img src="https://img.shields.io/github/v/release/AshokShau/SpTubeBot?style=for-the-badge&logo=github" alt="Release"/>
  </a>
  <br>
  <a href="https://goreportcard.com/report/github.com/AshokShau/SpTubeBot">
    <img src="https://goreportcard.com/badge/github.com/AshokShau/SpTubeBot?style=for-the-badge" alt="Go Report"/>
  </a>
  <a href="https://img.shields.io/github/go-mod/go-version/AshokShau/SpTubeBot">
    <img src="https://img.shields.io/github/go-mod/go-version/AshokShau/SpTubeBot?style=for-the-badge&logo=go" alt="Go Version"/>
  </a>
  <a href="https://github.com/AshokShau/SpTubeBot/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/License-MIT-blue?style=for-the-badge" alt="License"/>
  </a>
</p>

## 🌟 Features

### ⚡ Multi-Platform Support
- Spotify (tracks/albums/playlists)
- YouTube (videos/playlists)
- Apple Music (tracks/albums/playlists)
- SoundCloud  (tracks/playlists)

### 🤖 Telegram Integration
- Works in PM and groups
- Inline query support
- Fast response times
- High quality audio
- Seamless integration with Telegram groups

### 🛠 Technical
- Written in Go for high performance
- Docker-ready container
- Easy configuration

## 🚀 Quick Deployment

### Prerequisites
- Telegram Bot Token ([@BotFather](https://t.me/BotFather))
- API Key ([@FallenApiBot](https://t.me/FallenApiBot))
- Go 1.24+ or Docker

### Basic Installation
```bash
git clone https://github.com/AshokShau/SpTubeBot
cd SpTubeBot
cp sample.env .env
# Edit .env with your credentials
go build -o songBot
./songBot
```

### Docker Setup
```bash
docker build -t songbot .
docker run -d --name songbot --env-file .env songbot
```

## 📋 Command Reference

| Command            | Description              | Example                                       |
|--------------------|--------------------------|-----------------------------------------------|
| `/start`           | Welcome message          | `/start`                                      |
| `/spotify [query]` | Download from Spotify    | `/spotify https://open.spotify.com/track/...` |
| `/playlist [url]`  | Download entire playlist | `/playlist [playlist-url]`                    |
| `/ping`            | Check bot status         | `/ping`                                       |
| `/help`            | Show help message        | `/help`                                       |

## 🌐 Supported URL Formats (Just send the URL)
- Spotify: `open.spotify.com/track/...`
- YouTube: `youtube.com/watch?v=...`
- Apple Music: `music.apple.com/...`
- SoundCloud: `soundcloud.com/...`
- Instagram: `instagram.com/p/...` (Reels, Stories, and Posts)
- Pinterest: `pinterest.com/...` (Boards and Pins)

## 🆘 Support & Community
- [Report Issues](https://github.com/AshokShau/SpTubeBot/issues)
- [Telegram Support](https://t.me/FallenProjects)
- [Feature Requests](https://github.com/AshokShau/SpTubeBot/discussions)

## 📜 License
MIT Licensed - See [LICENSE](/LICENSE) for details.

---

<p align="center">
🎧 Enjoy unlimited music! Please ⭐ the repo if you find this useful.
</p>
