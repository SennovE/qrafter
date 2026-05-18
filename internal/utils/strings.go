package utils

import (
	"strings"
)

func QuoteWith(s, quote string) string {
	return quote + strings.ReplaceAll(s, quote, quote+quote) + quote
}
