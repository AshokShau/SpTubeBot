package commands

import (
	"fmt"
	"noinoi/internal/httpx"
	"os"
	"strings"

	"github.com/AshokShau/gotdbot"
)

const (
	MaxFileSize = 200 * 1024 * 1024 // 200 MB
)

func getArgs(m *gotdbot.Message) []string {
	text := m.Text()
	if text == "" {
		return nil
	}
	parts := strings.Fields(text)
	if len(parts) <= 1 {
		return nil
	}
	return parts[1:]
}

func catboxHandler(c *gotdbot.Client, m *gotdbot.Message) error {
	args := getArgs(m)
	userhash := globalConfig.CatboxUserhash

	// Owner only subcommands for album and file deletion
	if len(args) > 0 {
		if m.SenderID() != globalConfig.OwnerId {
			_, _ = m.ReplyText(c, "This command is only for the bot owner.", nil)
			return nil
		}

		if userhash == "" {
			_, _ = m.ReplyText(c, "Please set CATBOX_USERHASH", nil)
			return nil
		}

		switch args[0] {
		case "del":
			if len(args) < 2 {
				_, _ = m.ReplyText(c, "Usage: /catbox del <file1> <file2> ...", nil)
				return nil
			}
			err := httpx.DeleteCatboxFiles(args[1:], userhash)
			if err != nil {
				_, _ = m.ReplyText(c, fmt.Sprintf("Error deleting files: %v", err), nil)
			} else {
				_, _ = m.ReplyText(c, "Files successfully deleted.", nil)
			}
			return nil

		case "album":
			if len(args) < 2 {
				_, _ = m.ReplyText(c, "Usage: /catbox album <create|edit|add|remove|delete> ...", nil)
				return nil
			}
			return handleAlbumSubcommand(c, m, args[1:])
		}
	}

	if m.ReplyToMessageID() == 0 {
		_, _ = m.ReplyText(c, "Reply to a file, image, or video to upload it to Catbox.", nil)
		return nil
	}

	replyTo, err := m.GetRepliedMessage(c)
	if err != nil {
		_, _ = m.ReplyText(c, fmt.Sprintf("Failed to get replied message: %v", err), nil)
		return nil
	}

	fileId, fileSize, err := getFileIdAndSize(replyTo)
	if err != nil {
		_, _ = m.ReplyText(c, "Could not find a file to upload in the replied message.", nil)
		return nil
	}

	if fileSize > MaxFileSize {
		_, _ = m.ReplyText(c, "File is too large. Max size is 200MB.", nil)
		return nil
	}

	statusMsg, _ := m.ReplyText(c, "⏳ Downloading file...", nil)

	file, err := c.DownloadFile(fileId, 0, 0, 1, &gotdbot.DownloadFileOpts{Synchronous: true})
	if err != nil {
		_, _ = statusMsg.EditText(c, fmt.Sprintf("Failed to download file: %v", err), nil)
		return nil
	}
	defer os.Remove(file.Local.Path)

	_, _ = statusMsg.EditText(c, "⏳ Uploading to Catbox...", nil)

	url, err := httpx.UploadToCatbox(file.Local.Path, userhash)
	if err != nil {
		_, _ = statusMsg.EditText(c, fmt.Sprintf("Failed to upload to Catbox: %v", err), nil)
		return nil
	}

	button := &gotdbot.ReplyMarkupInlineKeyboard{
		Rows: [][]gotdbot.InlineKeyboardButton{
			{
				{
					Text: "View on Catbox",
					Type: &gotdbot.InlineKeyboardButtonTypeUrl{Url: url},
				},
				{
					Text: "Copy URL",
					Type: &gotdbot.InlineKeyboardButtonTypeCopyText{Text: url},
				},
			},
		},
	}

	_, _ = statusMsg.EditText(c, fmt.Sprintf("Uploaded successfully!\n\nURL: %s", url), &gotdbot.EditTextMessageOpts{
		DisableWebPagePreview: true,
		ReplyMarkup:           button,
	})

	return nil
}

func litterboxHandler(c *gotdbot.Client, m *gotdbot.Message) error {
	args := getArgs(m)

	if m.ReplyToMessageID() == 0 {
		_, _ = m.ReplyText(c, "Reply to a file, image, or video to upload it to Litterbox.", nil)
		return nil
	}

	timeStr := "24h"
	if len(args) > 0 {
		switch args[0] {
		case "1h", "12h", "24h", "72h":
			timeStr = args[0]
		default:
			_, _ = m.ReplyText(c, "Invalid time. Use 1h, 12h, 24h, or 72h.", nil)
			return nil
		}
	}

	replyTo, err := m.GetRepliedMessage(c)
	if err != nil {
		_, _ = m.ReplyText(c, fmt.Sprintf("Failed to get replied message: %v", err), nil)
		return nil
	}

	fileId, fileSize, err := getFileIdAndSize(replyTo)
	if err != nil {
		_, _ = m.ReplyText(c, "Could not find a file to upload in the replied message.", nil)
		return nil
	}

	if fileSize > MaxFileSize {
		_, _ = m.ReplyText(c, "File is too large. Max size is 200MB.", nil)
		return nil
	}

	statusMsg, _ := m.ReplyText(c, "⏳ Downloading file...", nil)

	file, err := c.DownloadFile(fileId, 0, 0, 1, &gotdbot.DownloadFileOpts{Synchronous: true})
	if err != nil {
		_, _ = statusMsg.EditText(c, fmt.Sprintf("Failed to download file: %v", err), nil)
		return nil
	}
	defer os.Remove(file.Local.Path)

	_, _ = statusMsg.EditText(c, fmt.Sprintf("⏳ Uploading to Litterbox (%s)...", timeStr), nil)

	url, err := httpx.UploadToLitterbox(file.Local.Path, timeStr)
	if err != nil {
		_, _ = statusMsg.EditText(c, fmt.Sprintf("Failed to upload to Litterbox: %v", err), nil)
		return nil
	}

	button := &gotdbot.ReplyMarkupInlineKeyboard{
		Rows: [][]gotdbot.InlineKeyboardButton{
			{
				{
					Text: "View on Catbox",
					Type: &gotdbot.InlineKeyboardButtonTypeUrl{Url: url},
				},
				{
					Text: "Copy URL",
					Type: &gotdbot.InlineKeyboardButtonTypeCopyText{Text: url},
				},
			},
		},
	}

	_, _ = statusMsg.EditText(c, fmt.Sprintf("Uploaded successfully (%s)!\n\nURL: %s", timeStr, url), &gotdbot.EditTextMessageOpts{
		DisableWebPagePreview: true,
		ReplyMarkup:           button,
	})

	return nil
}

func handleAlbumSubcommand(c *gotdbot.Client, m *gotdbot.Message, args []string) error {
	userhash := globalConfig.CatboxUserhash
	if userhash == "" {
		_, _ = m.ReplyText(c, "Please set CATBOX_USERHASH", nil)
		return nil
	}

	switch args[0] {
	case "create":
		// /catbox album create <title> | <desc> | <file1> <file2> ...
		fullArgs := strings.Join(args[1:], " ")
		parts := strings.Split(fullArgs, "|")
		if len(parts) < 3 {
			_, _ = m.ReplyText(c, "Usage: /catbox album create <title> | <desc> | <file1> <file2> ...", nil)
			return nil
		}
		title := strings.TrimSpace(parts[0])
		desc := strings.TrimSpace(parts[1])
		files := strings.Fields(strings.TrimSpace(parts[2]))
		url, err := httpx.CreateCatboxAlbum(title, desc, files, userhash)
		if err != nil {
			_, _ = m.ReplyText(c, fmt.Sprintf("Error creating album: %v", err), nil)
		} else {
			_, _ = m.ReplyText(c, fmt.Sprintf("Album created: %s", url), nil)
		}

	case "edit":
		// /catbox album edit <short> | <title> | <desc> | <file1> <file2> ...
		fullArgs := strings.Join(args[1:], " ")
		parts := strings.Split(fullArgs, "|")
		if len(parts) < 4 {
			_, _ = m.ReplyText(c, "Usage: /catbox album edit <short> | <title> | <desc> | <file1> <file2> ...", nil)
			return nil
		}
		short := strings.TrimSpace(parts[0])
		title := strings.TrimSpace(parts[1])
		desc := strings.TrimSpace(parts[2])
		files := strings.Fields(strings.TrimSpace(parts[3]))
		err := httpx.EditCatboxAlbum(short, title, desc, files, userhash)
		if err != nil {
			_, _ = m.ReplyText(c, fmt.Sprintf("Error editing album: %v", err), nil)
		} else {
			_, _ = m.ReplyText(c, "Album edited successfully.", nil)
		}

	case "add":
		if len(args) < 3 {
			_, _ = m.ReplyText(c, "Usage: /catbox album add <short> <file1> <file2> ...", nil)
			return nil
		}
		err := httpx.AddToCatboxAlbum(args[1], args[2:], userhash)
		if err != nil {
			_, _ = m.ReplyText(c, fmt.Sprintf("Error adding to album: %v", err), nil)
		} else {
			_, _ = m.ReplyText(c, "Files added to album.", nil)
		}

	case "remove":
		if len(args) < 3 {
			_, _ = m.ReplyText(c, "Usage: /catbox album remove <short> <file1> <file2> ...", nil)
			return nil
		}
		err := httpx.RemoveFromCatboxAlbum(args[1], args[2:], userhash)
		if err != nil {
			_, _ = m.ReplyText(c, fmt.Sprintf("Error removing from album: %v", err), nil)
		} else {
			_, _ = m.ReplyText(c, "Files removed from album.", nil)
		}

	case "delete":
		if len(args) < 2 {
			_, _ = m.ReplyText(c, "Usage: /catbox album delete <short>", nil)
			return nil
		}
		err := httpx.DeleteCatboxAlbum(args[1], userhash)
		if err != nil {
			_, _ = m.ReplyText(c, fmt.Sprintf("Error deleting album: %v", err), nil)
		} else {
			_, _ = m.ReplyText(c, "Album deleted successfully.", nil)
		}
	}
	return nil
}

func getFileIdAndSize(m *gotdbot.Message) (int32, int64, error) {
	if m.Content == nil {
		return 0, 0, fmt.Errorf("no content")
	}

	switch c := m.Content.(type) {
	case *gotdbot.MessagePhoto:
		if len(c.Photo.Sizes) > 0 {
			photo := c.Photo.Sizes[len(c.Photo.Sizes)-1]
			return photo.Photo.Id, photo.Photo.Size, nil
		}
	case *gotdbot.MessageVideo:
		return c.Video.Video.Id, c.Video.Video.Size, nil
	case *gotdbot.MessageAudio:
		return c.Audio.Audio.Id, c.Audio.Audio.Size, nil
	case *gotdbot.MessageDocument:
		return c.Document.Document.Id, c.Document.Document.Size, nil
	case *gotdbot.MessageAnimation:
		return c.Animation.Animation.Id, c.Animation.Animation.Size, nil
	case *gotdbot.MessageVoiceNote:
		return c.VoiceNote.Voice.Id, c.VoiceNote.Voice.Size, nil
	case *gotdbot.MessageVideoNote:
		return c.VideoNote.Video.Id, c.VideoNote.Video.Size, nil
	case *gotdbot.MessageSticker:
		return c.Sticker.Sticker.Id, c.Sticker.Sticker.Size, nil
	}

	return 0, 0, fmt.Errorf("unsupported content type %T", m.Content)
}
