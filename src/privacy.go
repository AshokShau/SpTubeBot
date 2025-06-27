package src

import (
	"fmt"
	"github.com/amarnathcjd/gogram/telegram"
)

// PrivacyHandle sends the bot's privacy policy to the user
func PrivacyHandle(m *telegram.NewMessage) error {
	bot := m.Client.Me()
	botName := bot.FirstName
	githubURL := "https://github.com/AshokShau/SpTubeBot"
	contactURL := "https://t.me/FallenProjects"

	privacyText := fmt.Sprintf(`
<b>Privacy Policy for %s</b>

<b>Last updated:</b> 8 June 2025

Thank you for using <b>@%s</b>. Your privacy is important to us. This Privacy Policy explains how we handle your information.

<b>1. Data Collection and Storage</b>
We do <b>not</b> collect, store, or share any user data.
- No message logging (text, commands, or requests)
- No user information storage (usernames, IDs, etc.)
- All processing happens in real-time with no persistent storage

<b>2. How We Work</b>
%s helps download songs from various platforms:
- Processes links/search queries
- Returns audio files directly to you
- Immediately discards all related data after delivery

<b>3. Third-Party Services</b>
We interact with these platforms (review their policies):
- YouTube
- Spotify
- Apple Music
- SoundCloud
%s <b>does not control</b> these services' data practices.

<b>4. Open Source Transparency</b>
The complete source code is available for review:
<a href="%s">%s</a>

<b>5. Security</b>
While no data is stored, we implement basic security measures for your protection during processing.

<b>6. Policy Updates</b>
We may update this policy occasionally. The "Last updated" date will reflect changes.

<b>7. Contact Us</b>
For questions or concerns:
<a href="%s">@FallenProjects</a> (Telegram)
or GitHub Issues
`,
		botName, botName, botName, botName,
		githubURL, githubURL,
		contactURL,
	)

	// Add a keyboard with quick links
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
