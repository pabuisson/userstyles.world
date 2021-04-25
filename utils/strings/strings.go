package strings

import (
	"regexp"
	"strings"
)

var slugRe = regexp.MustCompile(`[a-zA-Z0-9]+`)

func SlugifyURL(s string) string {
	// Extract valid characters.
	parts := slugRe.FindAllString(s, -1)

	// Join parts and make them lowercase.
	s = strings.Join(parts, "-")
	s = strings.ToLower(s)

	return s
}