# SongBot
A Telegram bot for downloading high-quality songs.


#### Api Key or Url ?
> [View Details](https://github.com/AshokShau/TgMusicBot?tab=readme-ov-file#facing-ip-ban-issues-from-youtube)

> [Api Docs](https://gist.github.com/AshokShau/7528cddc5b264035dee40523a44ff153)

## Features

- Download high-quality songs from Spotify
- Support for inline mode and private chats
- Ability to download songs from Spotify URLs
- Upload files directly to Telegram
- Docker support for easy deployment

## Installation

### 1. Install Go
Follow the official Go installation guide: [Go Installation Guide](https://golang.org/doc/install).

#### Easy Installation (Linux)
```shell
git clone https://github.com/udhos/update-golang dlgo && cd dlgo && sudo ./update-golang.sh && source /etc/profile.d/golang_path.sh
```
Exit the terminal and reopen it to verify the installation:
```shell
go version
```

### 2. Clone the Repository
```shell
git clone https://github.com/AshokShau/SpTubeBot && cd SpTubeBot
```

### 3. Set Up the Environment
Copy the sample environment file and edit it as needed:
```shell
cp sample.env .env && vi .env
```

### 4. Build the Project
```shell
go build -o songBot
```

### 5. Run the Bot
```shell
./songBot
```

## Docker Deployment

### 1. Build the Docker Image
```shell
docker build -t songbot .
```

### 2. Run the Bot Using Docker
```shell
docker run --env-file .env --name songbot -d songbot
```

## Usage

- Start the bot and send a song name or a Spotify URL to download the song.
- Use the `/spotify` command to download a song from Spotify in a group chat.

## License

This project is licensed under the MIT License. See the [LICENSE](/LICENSE) file for more information.
