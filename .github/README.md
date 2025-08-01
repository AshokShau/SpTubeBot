# 🎵 SongBot - Telegram Music Downloader Bot

A high-performance Telegram bot to download songs from **Spotify** and **YouTube** in 320kbps quality.

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
  <a href="https://github.com/AshokShau/SpTubeBot/blob/dev/LICENSE">
    <img src="https://img.shields.io/badge/License-MIT-blue?style=flat-square" alt="License"/>
  </a>
  <a href="https://t.me/FallenApiBot">
    <img src="https://img.shields.io/badge/API_Key-Required-important?style=flat-square" alt="API Key"/>
  </a>
</p>

---

## 🌟 Features

* 🎧 Download music in **320kbps** quality
* 🔗 Supports Multiple Platforms
* 📥 Works with tracks, albums, and playlists
* 🤖 Telegram inline support + command mode
* 💾 Built-in cache for faster responses
* 🐳 Docker-ready for seamless deployment

### Platforms
```
Stream, download, and enjoy music from your favorite platforms:
• Spotify
• YouTube
• SoundCloud
• Apple Music

🎥 Also supports media from:
• Instagram (Reels, Posts, Stories)
• Pinterest
• Facebook (Videos)
• TikTok
• Twitter
```

---

## 🚀 Quick Start

### ⚙️ Prerequisites

* Python 3.10+
* [Telegram Bot Token](https://t.me/BotFather)
* API Key via [@FallenApiBot](https://t.me/FallenApiBot)

---

### 🧑‍💻 Manual Setup

```bash
sudo apt-get install git python3-pip ffmpeg tmux -y
pip3 install uv

git clone https://github.com/AshokShau/SpTubeBot

cd SpTubeBot
uv venv

source .venv/bin/activate
uv pip install -e .

cp sample.env .env
nano .env

# Run the bot
start
```

---

### 🐳 Docker Deployment

```bash
# Build the image
docker build -t sp-tube-bot .

# Run the container (Make sure to create a .env file first)
docker run -d --name songbot --env-file .env sp-tube-bot
```

---

## 🆘 Support

Have questions or found a bug?

* Open an issue: [GitHub Issues](https://github.com/AshokShau/SpTubeBot/issues)
* Telegram: [@FallenProjects](https://t.me/FallenProjects)

---

## 📜 License

Licensed under the [MIT License](/LICENSE).

---

<p align="center">
  ❤️ Enjoy the music? Star the repo & share the bot with your friends!
</p>

---

