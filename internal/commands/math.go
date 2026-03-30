package commands

import (
	"fmt"
	"noinoi/internal/httpx"
	"strings"

	"github.com/AshokShau/gotdbot"
)

func mathHandler(c *gotdbot.Client, ctx *gotdbot.Context) error {
	m := ctx.EffectiveMessage
	args := m.GetText()
	parts := strings.SplitN(args, " ", 2)

	if len(parts) < 2 {
		_, _ = m.ReplyText(c, "<b>Usage:</b> <code>/math 2+2</code>", &gotdbot.SendTextMessageOpts{ParseMode: "HTML"})
		return nil
	}

	expression := parts[1]
	result, err := httpx.EvaluateExpression(expression)
	if err != nil {
		_, _ = m.ReplyText(c, fmt.Sprintf("<b>Error:</b> <code>%v</code>", err), &gotdbot.SendTextMessageOpts{ParseMode: "HTML"})
		return nil
	}

	response := fmt.Sprintf("<b>Expression:</b> <code>%s</code>\n<b>Result:</b> <code>%s</code>", expression, result)
	_, _ = m.ReplyText(c, response, &gotdbot.SendTextMessageOpts{ParseMode: "HTML"})
	return nil
}
