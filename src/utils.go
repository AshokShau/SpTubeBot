package src

import (
	"fmt"
	"github.com/amarnathcjd/gogram/telegram"
	"songBot/src/utils"
)

func prepareTrackMessageOptions(track *utils.SpotifyTrackDetails, file any, thumb []byte, pm *telegram.ProgressManager, caption string) telegram.SendOptions {
	opts := telegram.SendOptions{
		Media:   file,
		Caption: caption,
		Attributes: []telegram.DocumentAttribute{
			&telegram.DocumentAttributeFilename{FileName: fmt.Sprintf("%s.ogg", track.Name)},
			&telegram.DocumentAttributeAudio{Title: track.Name, Performer: track.Artist + " @SpTubeBot", Duration: int32(track.Duration)},
		},
		Spoiler:  true,
		MimeType: "audio/mpeg",
		ReplyMarkup: telegram.NewKeyboard().AddRow(
			telegram.Button.URL("🎵 Info", fmt.Sprintf("https://song.link/s/%s", track.TC)),
		).Build(),
	}

	if pm != nil {
		opts.ProgressManager = pm
	}

	if thumb != nil {
		opts.Thumb = thumb
	}

	return opts
}

func GoGramVersion(m *telegram.NewMessage) error {
	_, _ = m.Reply("GoGram Version: " + telegram.Version)
	return nil
}
