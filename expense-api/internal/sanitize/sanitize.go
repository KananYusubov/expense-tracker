package sanitize

import (
	"html"
	"strings"
)

// String trims whitespace, removes non-printable control characters,
// and escapes dangerous HTML characters to prevent XSS.
func String(s string) string {
	cleaned := strings.TrimSpace(strings.Map(func(r rune) rune {
		// Strip ASCII control characters below 32 (e.g. null bytes), except newline and tab
		if r < 32 && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, s))

	return html.EscapeString(cleaned)
}

// Category cleans and normalizes a category:
// trims, collapses multiple spaces into a single space, and converts to lowercase.
func Category(s string) string {
	cleaned := String(s)
	return strings.ToLower(strings.Join(strings.Fields(cleaned), " "))
}
