package src

import (
	"github.com/amarnathcjd/gogram/telegram"
	"songBot/src/utils"
)

// prepareTrackMessageOptions builds SendOptions to send an audio track.
func prepareTrackMessageOptions(
	track *utils.TrackInfo,
	file any,
	thumb []byte,
	pm *telegram.ProgressManager,
	caption string,
) telegram.SendOptions {
	opts := telegram.SendOptions{
		Media:    file,
		Caption:  caption,
		MimeType: "audio/mpeg",
		Spoiler:  true,
		Attributes: []telegram.DocumentAttribute{
			&telegram.DocumentAttributeAudio{
				Title:     track.Name,
				Performer: track.Artist + " @FallenProjects",
				Duration:  int32(track.Duration),
			},
		},
		ReplyMarkup: telegram.NewKeyboard().AddRow(
			telegram.Button.URL("🎧 Fᴀʟʟᴇɴ Pʀᴏᴊᴇᴄᴛꜱ", "https://t.me/FallenProjects"),
		).Build(),
	}

	if thumb != nil {
		opts.Thumb = thumb
	}
	if pm != nil {
		opts.ProgressManager = pm
	}

	return opts
}

// GoGramVersion responds with the current GoGram version.
func GoGramVersion(m *telegram.NewMessage) error {
	_, err := m.Reply("🤖 <b>GoGram Version:</b> " + telegram.Version)
	return err
}
