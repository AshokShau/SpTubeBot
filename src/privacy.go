package src

import (
	"fmt"

	"github.com/amarnathcjd/gogram/telegram"
)

// privacyHandle sends the bot's privacy policy to the user
func privacyHandle(m *telegram.NewMessage) error {
	bot := m.Client.Me()
	botName := bot.FirstName
	botUsername := bot.Username
	githubURL := "https://github.com/AshokShau/SpTubeBot"
	contactURL := "https://t.me/FallenProjects"

	privacyText := fmt.Sprintf(`
<b>🔐 Privacy Policy for %s</b>

<b>Last updated:</b> 17 July 2025

Thank you for using <b>@%s</b>. Your privacy is important to us. This policy explains how your data is handled.

<b>📌 1. What We Store</b>
- No usernames, messages, files, user id, chat id or queries are stored.
- in another words, we do not collect any data from you.
- We do not use any tracking or analytics services.

<b>⚙️ 2. How the Bot Works</b>
@%s helps you download songs from platforms like:
- YouTube, Spotify, Apple Music, SoundCloud
We process your requests in real time and send back the results. After processing, all temporary data is immediately discarded.

<b>📡 3. Third-Party Services</b>
This bot interacts with external services. Please refer to their respective privacy policies:
- YouTube
- Spotify
- Apple Music
- SoundCloud

No data is collected from these services.

<b>🔍 4. Open Source & Transparency</b>
You can review the full source code and deployment instructions here:
<a href="%s">%s</a>

<b>🛡️ 5. Security</b>
While we do not store sensitive data, basic protection is in place to keep the service stable and secure.

<b>📢 6. Changes to This Policy</b>
We may update this policy from time to time. The "Last updated" date above will always reflect the latest version.

<b>📬 7. Contact</b>
If you have any questions or concerns:
<a href="%s">@FallenProjects</a> (Telegram)
or open an issue on GitHub.
`,
		botName, botUsername, botUsername, githubURL, githubURL, contactURL,
	)

	keyboard := telegram.NewKeyboard().
		AddRow(
			telegram.Button.URL("📂 GitHub", githubURL),
			telegram.Button.URL("📩 Contact", contactURL),
		)

	_, err := m.Reply(privacyText, telegram.SendOptions{
		ReplyMarkup: keyboard.Build(),
	})
	return err
}
