package hw02unpackstring

import (
	"errors"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(s string) (string, error) {
	if s == "" {
		return "", nil
	}

	runes := []rune(s)
	if unicode.IsDigit(runes[0]) {
		return "", ErrInvalidString
	}

	for i := 1; i < len(runes); i++ {
		if unicode.IsDigit(runes[i]) && unicode.IsDigit(runes[i-1]) {
			return "", ErrInvalidString
		}
	}

	var b strings.Builder
	for i := 0; i < len(runes); i++ {
		if i+1 < len(runes) && unicode.IsDigit(runes[i+1]) {
			count := int(runes[i+1] - '0')
			b.WriteString(strings.Repeat(string(runes[i]), count))
			i++
		} else {
			b.WriteRune(runes[i])
		}
	}

	return b.String(), nil
}
