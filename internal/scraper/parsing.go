package scraper

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func parseSpanishWord(word string) string {
	if word == "" {
		return word
	}

	t := transform.Chain(
		norm.NFD,
		runes.Remove(runes.In(unicode.Mn)),
		norm.NFC,
	)

	parsedWord, _, err := transform.String(t, word)
	if err != nil {
		fmt.Println("Error parsing word:", err)
		return word
	}

	return strings.TrimSpace(strings.ToLower(parsedWord))
}
