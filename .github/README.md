# 🎵 SongBot - Telegram Music Downloader Bot

A high-performance Telegram bot for downloading songs from Spotify and YouTube in premium quality.

<p align="center">
  <a href="https://github.com/AshokShau/SpTubeBot/stargazers">
    <img src="https://img.shields.io/github/stars/AshokShau/SpTubeBot?style=flat-square&logo=github" alt="Stars"/>
  </a>
  <a href="https://github.com/AshokShau/SpTubeBot/network/members">
    <img src="https://img.shields.io/github/forks/AshokShau/SpTubeBot?style=flat-square&logo=github" alt="Forks"/>
  </a>
  <a href="https://github.com/AshokShau/SpTubeBot/releases">
    <img src="https://img.shields.io/github/v/release/AshokShau/SpTubeBot?style=flat-square" alt="Release"/>
  </a>

  <a href="https://goreportcard.com/report/github.com/AshokShau/SpTubeBot">
    <img src="https://goreportcard.com/badge/github.com/AshokShau/SpTubeBot?style=flat-square" alt="Go Report Card"/>
  </a>
  <a href="https://img.shields.io/github/go-mod/go-version/AshokShau/SpTubeBot">
    <img src="https://img.shields.io/github/go-mod/go-version/AshokShau/SpTubeBot?style=flat-square" alt="Go Version"/>
  </a>
  <a href="https://golang.org/">
    <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go" alt="Go Version"/>
  </a>
  <a href="https://github.com/AshokShau/SpTubeBot/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/License-MIT-blue?style=flat-square" alt="License"/>
  </a>
  <a href="https://t.me/FallenApiBot">
    <img src="https://img.shields.io/badge/API_Key-Required-important?style=flat-square" alt="API Key"/>
  </a>
</p>

## 🌟 Key Features

- 🎧 Download 320kbps quality songs  
- ⚡ Lightning-fast Spotify link processing  
- 📥 Supports single tracks, albums, and playlists  
- 🤖 Seamless Telegram integration (PM + Groups)  
- 🐳 Docker-ready for easy deployment  

## 🚀 Quick Start

### Prerequisites

- [Go 1.21+](https://golang.org/dl/)
- [Telegram Bot Token](https://t.me/BotFather)
- [Spotify/YouTube API Keys](https://t.me/FallenApiBot)

### Basic Installation

```bash
# Clone the repository
git clone https://github.com/AshokShau/SpTubeBot
cd SpTubeBot

# Configure environment
cp sample.env .env
nano .env  # Edit with your credentials

# Build & Run
go build -o songBot
./songBot
````

## 🛠 Advanced Setup

### Docker Deployment

```bash
docker build -t songbot .
docker run -d --name songbot --env-file .env songbot
```

## 📚 Usage Guide

| Command          | Description                |
| ---------------- | -------------------------- |
| `/start`         | Show welcome message       |
| `/spotify [url]` | Download from Spotify link |
| `/vid [url]`     | Download from YouTube link |
| `/help`          | Show command reference     |

**Inline Mode**: Type `@FallenSongBot` in any chat to search songs instantly!

## 🆘 Support

For issues and feature requests:

* [GitHub Issues](https://github.com/AshokShau/SpTubeBot/issues)
* Telegram Support: [@FallenProjects](https://t.me/FallenProjects)

## 📜 License

MIT License - See [LICENSE](/LICENSE) for full text.

---

<p align="center">
❤️ Enjoy the music! Support the project by starring the repo.
</p>
