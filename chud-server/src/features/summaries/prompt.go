package summaries

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

const PromptVersion = 2

const maxSummaryLength = 2000

const systemPrompt = `Jesteś komentatorem w grupowej aplikacji do pilnowania nawyków. Dostajesz fakty z jednego tygodnia w formacie JSON.

Napisz po polsku krótkie podsumowanie tygodnia: 3–5 zdań zwykłego tekstu, bez nagłówków, list i formatowania.

Zasady:
- Używaj wyłącznie imion, liczb i aktywności z danych. Niczego nie wymyślaj i nie licz od nowa.
- Punkty: zrobione +3, poza planem +1, wymówka 0, opuszczone −2. "ratePercent" to skuteczność.
- Wspomnij o liderze rankingu i ewentualnej zmianie względem "lastWeek".
- Jeśli są, wspomnij o idealnym tygodniu ("perfectWeek"), najczęściej opuszczanej aktywności, najlepszym dniu i najciekawszej wymówce.
- Jeśli "weekFinished" jest false, zaznacz, że tydzień jeszcze trwa i to stan na dziś.
- Ton luźny, lekko złośliwy, ale życzliwy.`

func userPrompt(factsJSON string) string {
	return "Fakty z tygodnia:\n" + factsJSON
}

var thinkBlock = regexp.MustCompile(`(?s)<think>.*?(</think>|$)`)

func cleanResponse(text string) string {
	text = strings.TrimSpace(thinkBlock.ReplaceAllString(text, ""))
	if end := strings.LastIndexAny(text, ".!?…"); end >= 0 {
		_, size := utf8.DecodeRuneInString(text[end:])
		text = text[:end+size]
	}
	runes := []rune(text)
	if len(runes) > maxSummaryLength {
		text = strings.TrimSpace(string(runes[:maxSummaryLength])) + "…"
	}
	return text
}
