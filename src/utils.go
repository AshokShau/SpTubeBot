package src

import (
	"fmt"
	"github.com/amarnathcjd/gogram/telegram"
	"os"
	"songBot/src/utils"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// prepareTrackMessageOptions builds SendOptions for sending an audio track.
func prepareTrackMessageOptions(file any, thumb any, track *utils.TrackInfo, progress *telegram.ProgressManager) telegram.SendOptions {
	return telegram.SendOptions{
		ProgressManager: progress,
		Media:           file,
		Thumb:           thumb,
		Attributes:      buildAudioAttributes(track),
		Caption:         buildTrackCaption(track),
		MimeType:        "audio/mpeg",
		ReplyMarkup: telegram.NewKeyboard().AddRow(
			telegram.Button.URL("🎧 Fᴀʟʟᴇɴ Pʀᴏᴊᴇᴄᴛꜱ", "https://t.me/FallenProjects"),
		).Build(),
	}
}

// buildTrackCaption returns the caption string for a Spotify track.
func buildTrackCaption(track *utils.TrackInfo) string {
	return fmt.Sprintf("<b>🎵 %s - %d</b>\n<b>Artist:</b> %s", track.Name, track.Year, track.Artist)
}

// buildAudioAttributes returns audio metadata for sending audio files.
func buildAudioAttributes(track *utils.TrackInfo) []telegram.DocumentAttribute {
	return []telegram.DocumentAttribute{
		&telegram.DocumentAttributeAudio{
			Title:     track.Name,
			Performer: track.Artist,
			Duration:  int32(track.Duration),
		},
	}
}

func clientSendEditedMessage(client *telegram.Client, msgID any, text string, opts *telegram.SendOptions) error {
	_, err := client.EditMessage(msgID, 0, text, opts)
	return err
}
