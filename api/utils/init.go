// Package utils contains utility functions
package utils

import (
	"crypto/rand"
	"math/big"
	"regexp"
	"strings"
	"unicode"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

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

func RandomAlphanumeric(length int) (string, error) {
	result := make([]byte, length)

	for i := range result {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[index.Int64()]
	}

	return string(result), nil
}
