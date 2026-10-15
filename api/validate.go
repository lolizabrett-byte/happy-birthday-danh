package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

const (
	maxWishName    = 30
	maxScoreName   = 20
	maxMessage     = 280
	maxScore       = 1000
	maxNewlinesMsg = 8
)

// validateName cleans (NFC, strip control/format chars, whitespace -> single spaces) and enforces 1..max runes.
func validateName(s string, max int) (string, error) {
	s = strings.Map(func(r rune) rune {
		switch {
		case r == '\t' || r == '\r' || r == '\n' || r == 0x2028 || r == 0x2029:
			return ' '
		case unicode.IsControl(r) || unicode.Is(unicode.Cf, r):
			return -1
		}
		return r
	}, s)
	s = strings.Join(strings.Fields(norm.NFC.String(s)), " ")
	n := utf8.RuneCountInString(s)
	if n == 0 {
		return "", errors.New("name is required")
	}
	if n > max {
		return "", fmt.Errorf("name must be at most %d characters", max)
	}
	return s, nil
}

// validateMessage cleans, limits blank lines (max one empty line in a row, max 8 newlines overall) and
// enforces 1..280 runes.
func validateMessage(s string) (string, error) {
	s = strings.NewReplacer("\r\n", "\n", "\r", "\n", "\xe2\x80\xa8", "\n", "\xe2\x80\xa9", "\n", "\t", " ").Replace(s) // U+2028, U+2029
	s = strings.Map(func(r rune) rune {
		if r != '\n' && (unicode.IsControl(r) || unicode.Is(unicode.Cf, r)) {
			return -1
		}
		return r
	}, s)
	lines := strings.Split(norm.NFC.String(s), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRightFunc(l, unicode.IsSpace)
	}
	s = strings.TrimSpace(strings.Join(lines, "\n"))
	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}
	s = capNewlines(s, maxNewlinesMsg)
	n := utf8.RuneCountInString(s)
	if n == 0 {
		return "", errors.New("message is required")
	}
	if n > maxMessage {
		return "", fmt.Errorf("message must be at most %d characters", maxMessage)
	}
	return s, nil
}

// capNewlines turns every newline beyond the first max into a space.
func capNewlines(s string, max int) string {
	count := 0
	return strings.Map(func(r rune) rune {
		if r == '\n' {
			if count++; count > max {
				return ' '
			}
		}
		return r
	}, s)
}

func validateScore(n int) error {
	if n < 0 || n > maxScore {
		return fmt.Errorf("score must be between 0 and %d", maxScore)
	}
	return nil
}
