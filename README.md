# SpTubeBot

SpTubeBot is a high-performance Telegram bot developed in Go using the `gotdbot` library. It provides advanced media downloading capabilities for various platforms, mathematical expression evaluation, and a unique bot cloning system.

## Features

- **Multi-Platform Media Downloader**: Supports downloading content from YouTube (Regular and Shorts), Instagram, TikTok, Pinterest, X (Twitter), Facebook, Threads, Twitch, Reddit, and Snapchat.
- **Music Downloader Support**: Automatically detects and downloads music from Spotify, SoundCloud, JioSaavn, and Tidal. It fetches track metadata and high-quality audio files.
- **YouTube Optimization**: Automatically detects YouTube links. Regular videos are delivered as M4A audio by default, while Shorts are delivered as MP4 videos.
- **Bot Cloning System**: Users can create their own instances of the bot by forwarding a message containing a bot token from @BotFather.
- **Mathematical Evaluation**: Evaluates complex mathematical expressions via a dedicated command.
- **Inline Query Support**: Allows users to search and share media directly in any chat using inline queries.
- **Dockerized Architecture**: Simplified deployment using Docker and Docker Compose.

## Prerequisites

- Go 1.25.6 or higher
- MongoDB instance
- `yt-dlp` (nightly builds recommended)
- `ffmpeg`
- Telegram API ID and API Hash
- Telegram Bot Token

## Environment Variables

The application requires the following environment variables. You can use the `sample.env` as a template.

| Variable    | Description                                           |
|-------------|-------------------------------------------------------|
| `API_ID`    | Telegram API ID from my.telegram.org                  |
| `API_HASH`  | Telegram API Hash from my.telegram.org                |
| `API_KEY`   | RapidAPI or internal service key for media extraction |
| `TOKEN`     | Primary bot token from @BotFather                     |
| `MONGO_URL` | MongoDB connection string                             |
| `API_URL`   | (Optional) Custom API endpoint for media extraction   |
| `OWNER_ID`  | (Optional) Telegram User ID of the bot owner          |

## Installation

### Using Docker (Recommended)

1. Clone the repository:
   ```bash
   git clone https://github.com/FallenProjects/SpTubeBot
   cd SpTubeBot
   ```

2. Create and configure your `.env` file:
   ```bash
   cp sample.env .env
   # Edit .env with your credentials
   ```

3. Build and start the container:
   ```bash
   docker-compose up -d --build
   ```

### Manual Installation

1. Ensure `yt-dlp` and `ffmpeg` are installed and available in your system's PATH.
2. Install the required `libtdjson` library:
   ```bash
   go run github.com/AshokShau/gotdbot/scripts/tools@latest
   ```
3. Build the application:
   ```bash
   go generate
   go build -o main .
   ```
4. Run the binary:
   ```bash
   ./main
   ```

## Bot Commands

- `/start` - Initialize interaction with the bot.
- `/ping` - Check system status, latency, and uptime.
- `/yt <url>` - Download a specific YouTube link as a high-quality video.
- `/math <expression>` - Evaluate a mathematical expression (e.g., `/math 2+2`).
- `/stop` - Stop a cloned instance and remove the token from the database (Owner only).

## Cloning Mechanism

To clone the bot:
1. Message @BotFather and create a new bot.
2. Forward the message containing the API token to the primary SpTubeBot instance.
3. The system will automatically register a new client and persist the configuration in MongoDB.

## Links

- **Repository**: [https://github.com/FallenProjects/SpTubeBot](https://github.com/FallenProjects/SpTubeBot)
- **Support**: [Join @FallenProjects](https://t.me/FallenProjects)
- **API**: [https://api.fallenapi.fun](https://api.fallenapi.fun)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
