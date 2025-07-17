package src

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

func uploadHandle(m *telegram.NewMessage) error {
	filename := m.Args()
	if filename == "" {
		_, _ = m.Reply("No filename provided")
		return nil
	}

	spoiler := false
	if strings.Contains(filename, "-s") {
		spoiler = true
		filename = strings.ReplaceAll(filename, "-s", "")
	}

	msg, _ := m.Reply("Uploading...")
	uploadStartTimestamp := time.Now()

	var pm = telegram.NewProgressManager(5)
	pm.Edit(telegram.MediaDownloadProgress(msg, pm))

	if _, err := m.RespondMedia(filename, telegram.MediaOptions{
		Spoiler:         spoiler,
		ProgressManager: pm,
	}); err != nil {
		msg.Edit("Error: " + err.Error())
		return nil
	} else {
		msg.Edit("Uploaded <code>" + filename + "</code> in <code>" + time.Since(uploadStartTimestamp).String() + "</code>")
	}

	return nil
}

func downloadHandle(m *telegram.NewMessage) error {
	if !m.IsReply() && m.Args() == "" {
		m.Reply("Reply to a file to download it")
		return nil
	}

	fn := m.Args()

	var r, msg *telegram.NewMessage
	if m.IsReply() {
		reply, err := m.GetReplyMessage()
		if err != nil {
			m.Reply("Error: " + err.Error())
			return nil
		}

		r = reply
		msg, _ = m.Reply("Downloading...")
	} else {
		reg := regexp.MustCompile(`t.me/(\w+)/(\d+)`)
		match := reg.FindStringSubmatch(m.Args())
		if len(match) != 3 || match[1] == "c" {
			// https://t.me/c/2183493392/8
			reg = regexp.MustCompile(`t.me/c/(\d+)/(\d+)`)
			match = reg.FindStringSubmatch(m.Args())
			if len(match) != 3 {
				m.Reply("Invalid link")
				return nil
			}

			id, err := strconv.Atoi(match[2])
			if err != nil {
				m.Reply("Invalid link: " + err.Error())
				return nil
			}

			chatID, err := strconv.Atoi(match[1])
			if err != nil {
				m.Reply("Invalid link: " + err.Error())
				return nil
			}

			msgX, err := m.Client.GetMessageByID(chatID, int32(id))
			if err != nil {
				m.Reply("Error: " + err.Error())
				return nil
			}
			r = msgX
			fn = r.File.Name
			msg, _ = m.Reply("Downloading... (from c " + strconv.Itoa(id) + ")")
		} else {
			username := match[1]
			id, err := strconv.Atoi(match[2])
			if err != nil {
				m.Reply("Invalid link")
				return nil
			}

			msgX, err := m.Client.GetMessageByID(username, int32(id))
			if err != nil {
				m.Reply("Error: " + err.Error())
				return nil
			}
			r = msgX
			fn = r.File.Name
			msg, _ = m.Reply("Downloading... (from " + username + " " + strconv.Itoa(id) + ")")
		}
	}

	uploadStartTimestamp := time.Now()

	var pm = telegram.NewProgressManager(5)
	pm.Edit(telegram.MediaDownloadProgress(msg, pm))

	if fi, err := r.Download(&telegram.DownloadOptions{
		ProgressManager: pm,
		FileName:        fn,
	}); err != nil {
		msg.Edit("Error: " + err.Error())
		return nil
	} else {
		msg.Edit("Downloaded <code>" + fi + "</code> in <code>" + time.Since(uploadStartTimestamp).String() + "</code>")
	}

	return nil
}
