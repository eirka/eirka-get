package utils

import (
	"strings"
	"unicode"
)

// FormatQuery removes MySQL special characters while preserving alphanumeric, spaces, and CJK characters
// This allows safe searching with full-text search using foreign language input
func FormatQuery(str string) string {
	return strings.Map(func(r rune) rune {
		// MySQL special characters that need to be filtered out
		if strings.ContainsRune(`"'+-@><()~*`, r) {
			return -1
		}

		// Keep alphanumeric, spaces, and valid characters
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) ||
			// Japanese Hiragana and Katakana
			(r >= 0x3040 && r <= 0x30FF) ||
			// CJK Unified Ideographs (common)
			(r >= 0x4E00 && r <= 0x9FFF) ||
			// CJK Unified Ideographs Extension A
			(r >= 0x3400 && r <= 0x4DBF) ||
			// Hangul (Korean)
			(r >= 0xAC00 && r <= 0xD7AF) ||
			// Fullwidth forms
			(r >= 0xFF00 && r <= 0xFFEF) {
			return r
		}

		// Filter out other potentially problematic characters
		return -1
	}, str)
}
