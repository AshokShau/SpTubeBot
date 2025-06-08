package src

import (
	"github.com/amarnathcjd/gogram/telegram"
)

const Privacy = `
<b>Privacy Policy for SpYtDlBot</b>

<b>Last updated:</b> 8 June 2025

Thank you for using <b>@SpYtDlBot</b>. Your privacy is important to us. This Privacy Policy outlines what information we do and do not collect when you use this Telegram bot.

<b>1. Data Collection and Storage</b><br>
We do <b>not collect, store, or share</b> any user data.<br>
- The bot does <b>not</b> log your messages, usernames, IDs, or song requests.<br>
- All processing is done in real time, and <b>no data is saved on the server</b> after your request is completed.

<b>2. Functionality</b><br>
SpYtDlBot helps users download songs from platforms like <b>Spotify</b> and <b>YouTube</b> via Telegram.<br>
- It processes links or search queries and returns audio files to the user.<br>
- The bot acts as a temporary bridge — once your song is delivered, the related data is discarded.

<b>3. Third-Party Services</b><br>
To function, SpYtDlBot fetches data from third-party platforms like:<br>
- <b>YouTube</b><br>
- <b>Spotify</b><br>
These services have their own privacy policies and data handling practices. SpYtDlBot <b>does not control or assume responsibility</b> for them.

<b>4. Open Source</b><br>
This bot is fully open source. You can review, audit, or contribute to the project on GitHub:<br>
<a href="https://github.com/AshokShau/SpTubeBot">https://github.com/AshokShau/SpTubeBot</a>

<b>5. Security</b><br>
Although no data is stored, we take basic steps to ensure your interaction with the bot is secure and private during processing.

<b>6. Changes to This Policy</b><br>
We may update this Privacy Policy if necessary. Any changes will be posted here with the updated date.

<b>7. Contact</b><br>
For any questions or concerns, feel free to reach out via GitHub Issues or Telegram:<br>
<a href="https://t.me/FallenProjects">@FallenProjects</a>
`

func PrivacyHandle(m *telegram.NewMessage) error {
	_, _ = m.Reply(Privacy)
	return nil
}
