package utils

import (
	"regexp"
	"strings"
	"unicode"
)

func Slugify(input string) string {
	input = strings.ToLower(input)
	input = strings.TrimSpace(input)

	var b strings.Builder
	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}

	slug := b.String()

	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")

	return strings.Trim(slug, "-")
}
