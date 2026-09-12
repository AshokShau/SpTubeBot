package commands

import (
	"fmt"

	"github.com/AshokShau/gotdbot"
)

func startHandler(c *gotdbot.Client, m *gotdbot.Message) error {
	username := c.Me.Usernames.EditableUsername
	text := fmt.Sprintf(`<h1>🚀 NoiNoi Bot</h1>

<p><b>Your Ultimate All-in-One Media Downloader &amp; Utility Bot!</b></p>

<blockquote>
Download videos, audio, tracks, and media effortlessly from Instagram, TikTok, YouTube, Spotify, SoundCloud, and more — all in one place.
</blockquote>

<h2>✨ Supported Platforms</h2>
<ul>
<li><b>YouTube &amp; Shorts</b>: High quality video &amp; audio extraction</li>
<li><b>Music</b>: Spotify, SoundCloud &amp; YouTube Music tracks</li>
<li><b>Social Media</b>: Instagram Reels, TikTok, Twitter/X, Pinterest, Snapchat, Facebook, Threads, Twitch, Reddit</li>
<li><b>File Hosting</b>: Instant uploads to Catbox &amp; Litterbox</li>
</ul>

<h2>💡 How It Works</h2>
<p>Simply send any supported link directly to this chat or group to get your download instantly!</p>

<details>
<summary><b>🛠️ Available Commands &amp; Utilities</b></summary>
<p>
• <code>/start</code> or <code>/help</code> — Show this help message<br/>
• <code>/yt &lt;link&gt;</code> — Download YouTube video in high quality<br/>
• <code>/math &lt;expr&gt;</code> — Calculate mathematical expressions<br/>
• <code>/catbox</code> or <code>/tgm</code> — Upload replied media or URL to Catbox.moe<br/>
• <code>/litterbox</code> — Upload replied media or URL to Litterbox<br/>
• <code>@%[1]s sound &lt;query&gt;</code> — Search and share sounds via inline query<br/>
• <code>/ping</code> — Check bot latency &amp; system uptime
</p>
</details>

<tg-button-row align="center">
  <tg-button type="url" style="primary" url="https://t.me/%[1]s?startgroup=true">➕ Add Me to Your Group</tg-button>
  <tg-button type="url" style="success" url="https://t.me/FallenProjects">📢 Updates Channel</tg-button>
</tg-button-row>
<tg-button-row align="center">
  <tg-button type="switch_inline_query" query="CID">🔊 Sound Search</tg-button>
</tg-button-row>`, username)

	richMessage := &gotdbot.InputRichMessage{
		DetectAutomaticBlocks: true,
		Source: &gotdbot.RichMessageSourceHtml{
			Text: text,
		},
	}

	_, err := m.ReplyRichMessage(c, richMessage, nil)
	return err
}

func cloneHandler(c *gotdbot.Client, m *gotdbot.Message) error {
	return gotdbot.EndGroups
}
