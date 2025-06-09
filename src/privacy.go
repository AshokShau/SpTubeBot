package src

import (
	"fmt"
	"github.com/amarnathcjd/gogram/telegram"
)

func PrivacyHandle(m *telegram.NewMessage) error {
	botName := m.Client.Me().FirstName

	privacyText := fmt.Sprintf(`
<b>Privacy Policy for %s</b>

<b>Last updated:</b> 8 June 2025

Thank you for using <b>@%s</b>. Your privacy is important to us. This Privacy Policy outlines what information we do and do not collect when you use this Telegram bot.

<b>1. Data Collection and Storage</b>
We do <b>not collect, store, or share</b> any user data.
- The bot does <b>not</b> log your messages, usernames, IDs, or song requests.
- All processing is done in real time, and <b>no data is saved on the server</b> after your request is completed.

<b>2. Functionality</b>
%s helps users download songs from platforms like <b>Spotify</b> and <b>YouTube</b> via Telegram.
- It processes links or search queries and returns audio files to the user.
- The bot acts as a temporary bridge — once your song is delivered, the related data is discarded.

<b>3. Third-Party Services</b>
To function, %s fetches data from third-party platforms like:
- <b>YouTube</b>
- <b>Spotify</b>
These services have their own privacy policies and data handling practices. %s <b>does not control or assume responsibility</b> for them.

<b>4. Open Source</b>
This bot is fully open source. You can review, audit, or contribute to the project on GitHub: <a href="https://github.com/AshokShau/SpTubeBot">https://github.com/AshokShau/SpTubeBot</a>

<b>5. Security</b>
Although no data is stored, we take basic steps to ensure your interaction with the bot is secure and private during processing.

<b>6. Changes to This Policy</b>
We may update this Privacy Policy if necessary. Any changes will be posted here with the updated date.

<b>7. Contact</b><br>
For any questions or concerns, feel free to reach out via GitHub Issues or Telegram:
<a href="https://t.me/FallenProjects">@FallenProjects</a>
`, botName, botName, botName, botName, botName)

	_, _ = m.Reply(privacyText)
	return nil
}
