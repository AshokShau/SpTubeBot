package commands

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"noinoi/internal/httpx"
	"strconv"
	"strings"
	"sync"
	"time"

	td "github.com/AshokShau/gotdbot"
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

func handleInlineQuery(c *td.Client, ctx *td.Context) error {
	iq := ctx.Update.UpdateNewInlineQuery
	query := strings.TrimSpace(iq.Query)
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

	var results []td.InputInlineQueryResult
	urlHash := setCachedURL(targetUrl)
	caption := "Join @FallenProjects"

	mediaList := getAllMedia(snapData)
	if len(mediaList) == 0 {
		return nil
	}

	for i, media := range mediaList {
		markup := createNavigationMarkup(urlHash, i, len(mediaList))
		id := fmt.Sprintf("snap_%s_%d", urlHash, i)

		var result td.InputInlineQueryResult
		thumb := media.Thumbnail
		if thumb == "" {
			thumb = "https://placehold.co/200x200/png?text=No+Thumbnail"
		}

		if media.Type == "video" || media.Type == "animation" {
			result = &td.InputInlineQueryResultVideo{
				Id:           id,
				Title:        snapData.Title,
				VideoUrl:     media.URL,
				MimeType:     "video/mp4",
				ThumbnailUrl: thumb,
				ReplyMarkup:  markup,
				InputMessageContent: &td.InputMessageVideo{
					Video: &td.InputFileRemote{Id: media.URL},
					Caption: &td.FormattedText{
						Text: caption,
					},
				},
			}
		} else {
			result = &td.InputInlineQueryResultPhoto{
				Id:           id,
				Title:        snapData.Title,
				PhotoUrl:     media.URL,
				ThumbnailUrl: thumb,
				ReplyMarkup:  markup,
				InputMessageContent: &td.InputMessagePhoto{
					Photo: &td.InputFileRemote{Id: media.URL},
					Caption: &td.FormattedText{
						Text: caption,
					},
				},
			}
		}
		results = append(results, result)
	}

	return c.AnswerInlineQuery(0, iq.Id, "", results, nil)
}

func handleInlineCallbackQuery(c *td.Client, ctx *td.Context) error {
	icq := ctx.Update.UpdateNewInlineCallbackQuery
	if isRateLimited(icq.SenderUserId) {
		_ = c.AnswerCallbackQuery(0, icq.Id, "Slow down! Don't spam.", "", &td.AnswerCallbackQueryOpts{ShowAlert: true})
		return nil
	}

	dataPayload, ok := icq.Payload.(*td.CallbackQueryPayloadData)
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

	var content td.InputMessageContent
	if media.Type == "video" || media.Type == "animation" {
		content = &td.InputMessageVideo{
			Video: &td.InputFileRemote{Id: media.URL},
			Caption: &td.FormattedText{
				Text: caption,
			},
		}
	} else {
		content = &td.InputMessagePhoto{
			Photo: &td.InputFileRemote{Id: media.URL},
			Caption: &td.FormattedText{
				Text: caption,
			},
		}
	}

	err = c.EditInlineMessageMedia(icq.InlineMessageId, content, &td.EditInlineMessageMediaOpts{
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

func createNavigationMarkup(urlHash string, currentIndex, total int) *td.ReplyMarkupInlineKeyboard {
	if total <= 1 {
		return nil
	}

	var buttons []td.InlineKeyboardButton
	prevIndex := (currentIndex - 1 + total) % total
	nextIndex := (currentIndex + 1) % total

	buttons = append(buttons, td.InlineKeyboardButton{
		Text: "⬅️ Previous",
		Type: &td.InlineKeyboardButtonTypeCallback{
			Data: []byte(fmt.Sprintf("sn_%s_%d", urlHash, prevIndex)),
		},
	})

	buttons = append(buttons, td.InlineKeyboardButton{
		Text: fmt.Sprintf("%d / %d", currentIndex+1, total),
		Type: &td.InlineKeyboardButtonTypeCallback{
			Data: []byte("none"),
		},
	})

	buttons = append(buttons, td.InlineKeyboardButton{
		Text: "Next ➡️",
		Type: &td.InlineKeyboardButtonTypeCallback{
			Data: []byte(fmt.Sprintf("sn_%s_%d", urlHash, nextIndex)),
		},
	})

	return &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{buttons},
	}
}
