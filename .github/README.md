# 🎵 SpTubeBot - Telegram Music & Media Downloader Bot

A powerful Telegram bot that lets you download music and media from various platforms in high quality.

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

## 🚀 Features

- 🎧 Download music in 320kbps quality
- 🔗 Supports multiple platforms:
  - Music: Spotify, YouTube, SoundCloud, Apple Music
  - Media: Instagram, Pinterest, Facebook, TikTok, Twitter, Threads
- 📥 Works with tracks, albums, and playlists
- 🤖 Telegram inline support and command mode
- 💾 Built-in cache for faster responses
- 🐳 Docker-ready for easy deployment
- 🔐 Secure environment-based configuration

## 📋 Prerequisites

Before you begin, make sure you have:

1. Python 3.10 or higher
2. A Telegram Bot Token from @BotFather
3. An API Key from @FallenApiBot
4. Required system dependencies:
   - ffmpeg (for audio processing)
   - tmux (for process management)

## 🛠️ Installation

### Option 1: Manual Installation

```bash
# 1. Install system dependencies
sudo apt-get install git python3-pip ffmpeg vorbis-tools tmux -y

# 2. Install uv (Python package manager)
pip3 install uv

# 3. Clone the repository
git clone https://github.com/AshokShau/SpTubeBot

cd SpTubeBot

# 4. Create virtual environment
uv venv

# 5. Activate virtual environment
source .venv/bin/activate

# 6. Install dependencies
uv pip install -e .

# 7. Copy and edit environment file
cp sample.env .env
nano .env

# 8. Run the bot
start
```

### Option 2: Docker Deployment (Recommended)

```bash
# 1. Build the Docker image
docker build -t sp-tube-bot .

# 2. Run the container (Make sure to create a .env file first)
docker run -d --name songbot --env-file .env sp-tube-bot
```

## 📝 Environment Variables

Create a `.env` file with the following variables:

```env
API_ID=your_api_id
API_HASH=your_api_hash
TOKEN=your_telegram_bot_token
API_KEY=your_fallen_api_key # Get from @FallenApiBot
API_URL=https://tgmusic.fallenapi.fun
DOWNLOAD_PATH=database
LOGGER_ID=-1002434755494
```

## 🤖 Using the Bot

1. Start the bot using either manual or Docker deployment
2. Search for the bot in Telegram and start a chat
3. You can use the bot in two ways:
   - Inline mode: Type `@your_bot_name` in any chat and search for music
   - Command mode: Send commands directly to the bot

Available commands:
- `/start` - Start the bot
- `/help` - Get help and command list
- `/song` - Download a song
- `/playlist` - Download a playlist
- Just send a link to the bot and it will download the media


## 📝 License

This project is licensed under the MIT License - see the [LICENSE](/LICENSE) file for details.

## 👥 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 🙏 Support

If you encounter any issues or have questions:

1. Check the [issues](https://github.com/AshokShau/SpTubeBot/issues) page
2. Create a new issue if your problem isn't listed
3. For general questions, you can also message @AshokShau on Telegram

## 📝 Note

This bot is intended for personal use and educational purposes. Please respect copyright laws and terms of service when using this bot.
