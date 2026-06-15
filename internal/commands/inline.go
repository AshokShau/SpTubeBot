package commands

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log/slog"
	"noinoi/internal/httpx"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AshokShau/gotdbot"
)

var (
	urlCache = make(map[string]string)
	cacheMu  sync.RWMutex

	rateLimit = make(map[int64]int64)
	limitMu   sync.Mutex
)

func isRateLimited(userId int64) bool {
	limitMu.Lock()
	defer limitMu.Unlock()

	now := time.Now().Unix()
	last, ok := rateLimit[userId]
	if ok && now-last < 1 {
		return true
	}

	rateLimit[userId] = now
	return false
}

func getCachedURL(hash string) string {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	return urlCache[hash]
}

func setCachedURL(url string) string {
	hash := md5.Sum([]byte(url))
	hashStr := hex.EncodeToString(hash[:8])
	cacheMu.Lock()
	defer cacheMu.Unlock()
	urlCache[hashStr] = url
	return hashStr
}

func handleInlineQuery(c *gotdbot.Client, iq *gotdbot.UpdateNewInlineQuery) error {
	query := strings.TrimSpace(iq.Query)

	isUrl := strings.HasPrefix(query, "http://") || strings.HasPrefix(query, "https://")

	if query == "" || !isUrl {
		instants, err := httpx.GetInstants(query)
		if err != nil || len(instants) == 0 {
			slog.Warn("Failed to fetch instants", "error", err, "query", query)
			return nil
		}

		var results []gotdbot.InputInlineQueryResult
		for i, inst := range instants {
			results = append(results, &gotdbot.InputInlineQueryResultAudio{
				Id:       fmt.Sprintf("instant_%d", i),
				Title:    inst.Name,
				AudioUrl: inst.URL,
				InputMessageContent: &gotdbot.InputMessageAudio{
					Audio: &gotdbot.InputAudio{
						Audio: &gotdbot.InputFileRemote{Id: inst.URL},
					},
					Caption: &gotdbot.FormattedText{
						Text: "Join @FallenProjects",
					},
				},
			})
		}
		return c.AnswerInlineQuery(0, iq.Id, "", results, nil)
	}

	var targetUrl string
	for _, pattern := range httpx.SnapPatterns {
		if pattern.MatchString(query) {
			targetUrl = pattern.FindString(query)
			break
		}
	}

	if targetUrl == "" {
		return nil
	}

	snapData, err := httpx.GetSnap(targetUrl)
	if err != nil {
		return nil
	}

	var results []gotdbot.InputInlineQueryResult
	urlHash := setCachedURL(targetUrl)
	caption := "Join @FallenProjects"

	mediaList := getAllMedia(snapData)
	if len(mediaList) == 0 {
		return nil
	}

	totalItems := len(mediaList)
	if len(mediaList) > 50 {
		mediaList = mediaList[:50]
	}

	for i, media := range mediaList {
		markup := createNavigationMarkup(urlHash, i, totalItems)
		id := fmt.Sprintf("snap_%s_%d", urlHash, i)

		var result gotdbot.InputInlineQueryResult
		thumb := media.Thumbnail
		if thumb == "" {
			thumb = "https://placehold.co/200x200/png?text=No+Thumbnail"
		}

		if media.Type == "video" || media.Type == "animation" {
			result = &gotdbot.InputInlineQueryResultVideo{
				Id:           id,
				Title:        snapData.Title,
				VideoUrl:     media.URL,
				MimeType:     "video/mp4",
				ThumbnailUrl: thumb,
				ReplyMarkup:  markup,
				InputMessageContent: &gotdbot.InputMessageVideo{
					Video: &gotdbot.InputVideo{
						Video: &gotdbot.InputFileRemote{Id: media.URL},
					},
					Caption: &gotdbot.FormattedText{
						Text: caption,
					},
				},
			}
		} else {
			result = &gotdbot.InputInlineQueryResultPhoto{
				Id:           id,
				Title:        snapData.Title,
				PhotoUrl:     media.URL,
				ThumbnailUrl: thumb,
				ReplyMarkup:  markup,
				InputMessageContent: &gotdbot.InputMessagePhoto{
					Photo: &gotdbot.InputPhoto{
						Photo: &gotdbot.InputFileRemote{Id: media.URL},
					},
					Caption: &gotdbot.FormattedText{
						Text: caption,
					},
				},
			}
		}
		results = append(results, result)
	}

	return c.AnswerInlineQuery(0, iq.Id, "", results, nil)
}

func handleGuestQuery(c *gotdbot.Client, u *gotdbot.UpdateNewGuestQuery) error {
	var query string
	if len(u.ReferenceMessages) > 0 {
		query = u.ReferenceMessages[0].GetText()
	} else if u.Message != nil {
		query = u.Message.GetText()
	}

	if query == "" {
		return nil
	}

	var targetUrl string
	for _, pattern := range httpx.SnapPatterns {
		if pattern.MatchString(query) {
			targetUrl = pattern.FindString(query)
			break
		}
	}

	if targetUrl == "" {
		return nil
	}

	snapData, err := httpx.GetSnap(targetUrl)
	if err != nil {
		return nil
	}

	mediaList := getAllMedia(snapData)
	if len(mediaList) == 0 {
		return nil
	}

	urlHash := setCachedURL(targetUrl)
	caption := "Join @FallenProjects"

	media := mediaList[0]
	markup := createNavigationMarkup(urlHash, 0, len(mediaList))
	id := fmt.Sprintf("snap_%s_%d", urlHash, 0)
	thumb := media.Thumbnail
	if thumb == "" {
		thumb = "https://placehold.co/200x200/png?text=No+Thumbnail"
	}

	var result gotdbot.InputInlineQueryResult
	if media.Type == "video" || media.Type == "animation" {
		result = &gotdbot.InputInlineQueryResultVideo{
			Id:           id,
			Title:        snapData.Title,
			VideoUrl:     media.URL,
			MimeType:     "video/mp4",
			ThumbnailUrl: thumb,
			ReplyMarkup:  markup,
			InputMessageContent: &gotdbot.InputMessageVideo{
				Video: &gotdbot.InputVideo{
					Video: &gotdbot.InputFileRemote{Id: media.URL},
				},
				Caption: &gotdbot.FormattedText{
					Text: caption,
				},
			},
		}
	} else {
		result = &gotdbot.InputInlineQueryResultPhoto{
			Id:           id,
			Title:        snapData.Title,
			PhotoUrl:     media.URL,
			ThumbnailUrl: thumb,
			ReplyMarkup:  markup,
			InputMessageContent: &gotdbot.InputMessagePhoto{
				Photo: &gotdbot.InputPhoto{
					Photo: &gotdbot.InputFileRemote{Id: media.URL},
				},
				Caption: &gotdbot.FormattedText{
					Text: caption,
				},
			},
		}
	}

	_, err = c.AnswerGuestQuery(u.Id, result)
	if err != nil {
		c.Logger.Error("Failed to answer guest query", "error", err)
		errorResult := &gotdbot.InputInlineQueryResultArticle{
			Id:    strconv.FormatInt(time.Now().UnixNano(), 10),
			Title: "Error occurred",
			InputMessageContent: &gotdbot.InputMessageText{
				Text: &gotdbot.FormattedText{
					Text: fmt.Sprintf("An error occurred while processing your request: %v", err),
				},
			},
			Description: err.Error(),
		}
		_, _ = c.AnswerGuestQuery(u.Id, errorResult)
		return err
	}

	return nil
}

func handleInlineCallbackQuery(c *gotdbot.Client, icq *gotdbot.UpdateNewInlineCallbackQuery) error {
	if isRateLimited(icq.SenderUserId) {
		_ = c.AnswerCallbackQuery(0, icq.Id, "Slow down! Don't spam.", "", &gotdbot.AnswerCallbackQueryOpts{ShowAlert: true})
		return nil
	}

	dataPayload, ok := icq.Payload.(*gotdbot.CallbackQueryPayloadData)
	if !ok {
		return nil
	}

	data := string(dataPayload.Data)
	if !strings.HasPrefix(data, "sn_") {
		return nil
	}

	parts := strings.Split(data, "_")
	if len(parts) != 3 {
		return nil
	}

	urlHash := parts[1]
	index, _ := strconv.Atoi(parts[2])

	targetUrl := getCachedURL(urlHash)
	if targetUrl == "" {
		_ = c.AnswerCallbackQuery(0, icq.Id, "Session expired, please search again.", "", nil)
		return nil
	}

	snapData, err := httpx.GetSnap(targetUrl)
	if err != nil {
		_ = c.AnswerCallbackQuery(0, icq.Id, "Error fetching data.", "", nil)
		return nil
	}

	mediaList := getAllMedia(snapData)
	if index < 0 || index >= len(mediaList) {
		return nil
	}

	media := mediaList[index]
	markup := createNavigationMarkup(urlHash, index, len(mediaList))
	caption := "Join @FallenProjects"

	var content gotdbot.InputMessageContent
	if media.Type == "video" || media.Type == "animation" {
		content = &gotdbot.InputMessageVideo{
			Video: &gotdbot.InputVideo{
				Video: &gotdbot.InputFileRemote{Id: media.URL},
			},
			Caption: &gotdbot.FormattedText{
				Text: caption,
			},
		}
	} else {
		content = &gotdbot.InputMessagePhoto{
			Photo: &gotdbot.InputPhoto{
				Photo: &gotdbot.InputFileRemote{Id: media.URL},
			},
			Caption: &gotdbot.FormattedText{
				Text: caption,
			},
		}
	}

	err = c.EditInlineMessageMedia(icq.InlineMessageId, content, &gotdbot.EditInlineMessageMediaOpts{
		ReplyMarkup: markup,
	})
	if err != nil {
		_ = c.AnswerCallbackQuery(0, icq.Id, "Failed to update media.", "", nil)
	} else {
		_ = c.AnswerCallbackQuery(0, icq.Id, "", "", nil)
	}

	return nil
}

type mediaItem struct {
	URL       string
	Type      string
	Thumbnail string
}

func getAllMedia(snapData *httpx.SnapResponse) []mediaItem {
	var items []mediaItem
	for _, img := range snapData.Images {
		items = append(items, mediaItem{URL: img, Type: "photo", Thumbnail: img})
	}
	for _, vid := range snapData.Videos {
		items = append(items, mediaItem{URL: vid.URL, Type: "video", Thumbnail: vid.Thumbnail})
	}
	return items
}

func createNavigationMarkup(urlHash string, currentIndex, total int) *gotdbot.ReplyMarkupInlineKeyboard {
	if total <= 1 {
		return nil
	}

	var buttons []gotdbot.InlineKeyboardButton
	prevIndex := (currentIndex - 1 + total) % total
	nextIndex := (currentIndex + 1) % total

	buttons = append(buttons, gotdbot.InlineKeyboardButton{
		Text: "⬅️ Previous",
		Type: &gotdbot.InlineKeyboardButtonTypeCallback{
			Data: []byte(fmt.Sprintf("sn_%s_%d", urlHash, prevIndex)),
		},
	})

	buttons = append(buttons, gotdbot.InlineKeyboardButton{
		Text: fmt.Sprintf("%d / %d", currentIndex+1, total),
		Type: &gotdbot.InlineKeyboardButtonTypeCallback{
			Data: []byte("none"),
		},
	})

	buttons = append(buttons, gotdbot.InlineKeyboardButton{
		Text: "Next ➡️",
		Type: &gotdbot.InlineKeyboardButtonTypeCallback{
			Data: []byte(fmt.Sprintf("sn_%s_%d", urlHash, nextIndex)),
		},
	})

	return &gotdbot.ReplyMarkupInlineKeyboard{
		Rows: [][]gotdbot.InlineKeyboardButton{buttons},
	}
}
