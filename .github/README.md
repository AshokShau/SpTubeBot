# 🎵 SongBot Installation Guide

A Telegram bot for downloading high-quality songs from Spotify.

> 🔑 Need an API Key or URL?
> Talk to [@FallenApiBot](https://t.me/FallenApiBot)

---

## 💡 Features

* Download high-quality Spotify songs
* Supports inline mode & private chats
* Accepts Spotify song, album, or playlist URLs
* Uploads songs directly to Telegram
* Easy setup with or without Docker

---

# ⚙️ Installation Steps

## **1. Install Go (Golang)**

### ✅ Easiest Method for Everyone:

#### **🖥️ Windows / macOS:**

* Go to: [https://golang.org/doc/install](https://golang.org/doc/install)
* Download the installer and run it (just like any other app)
* After installing, open a terminal and check it's working:

  ```bash
  go version
  ```

#### **🐧 Linux (1-Line Install):**

Open your terminal and run:

```bash
git clone https://github.com/udhos/update-golang dlgo && cd dlgo && sudo ./update-golang.sh && source /etc/profile.d/golang_path.sh
```

Then close and reopen your terminal, and check:

```bash
go version
```

---

## **2. Download the Bot Code**

In your terminal, run:

```bash
git clone https://github.com/AshokShau/SpTubeBot && cd SpTubeBot
```

This will download the code and move you into the bot's folder.

---

## **3. Configure the Bot**

### 🔧 Setup the Environment File

1. Copy the example config:

   ```bash
   cp sample.env .env
   ```

2. Edit the `.env` file:

    * On **Windows**: Open it in Notepad or VS Code
    * On **Linux/Mac**:

      ```bash
      nano .env
      ```

      (Edit the values, then press `Ctrl + X`, `Y`, and Enter to save)

---

## **4. Run the Bot**

### 🟢 Option 1: Run Without Docker (Simple)

```bash
go build -o songBot
./songBot
```

### 🐳 Option 2: Run With Docker (Clean Setup)

```bash
docker build -t songbot .
docker run --env-file .env --name songbot -d songbot
```

---

## ✅ You’re Done! Start Using the Bot

* Open your bot on Telegram
* Send a song name or a Spotify link
* You can also use `/spotify` in group chats

---

## 🪪 License

This project uses the **MIT License**.
See the [LICENSE](/LICENSE) file for details.

---
