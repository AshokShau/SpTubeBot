package commands

import (
	"fmt"
	"runtime"
	"time"

	"github.com/AshokShau/gotdbot"
)

func pingHandler(c *gotdbot.Client, ctx *gotdbot.Context) error {
	m := ctx.EffectiveMessage
	start := time.Now()

	msg, err := m.ReplyText(c, "Checking status...", nil)
	if err != nil {
		return nil
	}

	latency := time.Since(start).Milliseconds()
	uptime := time.Since(startTime).Truncate(time.Second)

	response := fmt.Sprintf(
		"<b>System Status</b>\n\n"+
			"<b>Latency:</b> <code>%d ms</code>\n"+
			"<b>Uptime:</b> <code>%s</code>\n"+
			"<b>Go Routines:</b> <code>%d</code>\n",
		latency, uptime, runtime.NumGoroutine(),
	)

	_, _ = msg.EditText(c, response, &gotdbot.EditTextMessageOpts{ParseMode: "HTML"})
	return nil
}
