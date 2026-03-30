package commands

import td "github.com/AshokShau/gotdbot"

func getUrl(m *td.Message) string {
	text := m.GetText()
	if text == "" {
		return ""
	}

	entities := m.GetEntities()
	if entities == nil || len(entities) == 0 {
		return ""
	}

	for _, entity := range entities {
		switch t := entity.Type.(type) {

		case *td.TextEntityTypeUrl:
			start := entity.Offset
			end := entity.Offset + entity.Length
			if int(end) <= len(text) {
				return text[start:end]
			}

		case *td.TextEntityTypeTextUrl:
			return t.Url
		}
	}

	return ""
}
