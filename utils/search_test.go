package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatQuery(t *testing.T) {
	// Test with various special characters
	t.Run("Special characters", func(t *testing.T) {
		input := "test@char+act-er\"s'~*()><"
		expected := "testcharacters"
		result := FormatQuery(input)
		assert.Equal(t, expected, result, "Special characters should be removed")
	})

	// Test with a clean string
	t.Run("Clean string", func(t *testing.T) {
		input := "cleanstring"
		expected := "cleanstring"
		result := FormatQuery(input)
		assert.Equal(t, expected, result, "Clean string should remain unchanged")
	})

	// Test with Japanese characters
	t.Run("Japanese characters", func(t *testing.T) {
		// Hiragana, Katakana, and Kanji
		input := "こんにちは+世界*ジャパン"
		expected := "こんにちは世界ジャパン"
		result := FormatQuery(input)
		assert.Equal(t, expected, result, "Japanese characters should be preserved while special characters are removed")
	})

	// Test with Chinese characters
	t.Run("Chinese characters", func(t *testing.T) {
		input := "你好+世界*中国"
		expected := "你好世界中国"
		result := FormatQuery(input)
		assert.Equal(t, expected, result, "Chinese characters should be preserved while special characters are removed")
	})

	// Test with Korean characters
	t.Run("Korean characters", func(t *testing.T) {
		input := "안녕하세요+세계*한국"
		expected := "안녕하세요세계한국"
		result := FormatQuery(input)
		assert.Equal(t, expected, result, "Korean characters should be preserved while special characters are removed")
	})

	// Test with mixed character sets
	t.Run("Mixed character sets", func(t *testing.T) {
		input := "Hello世界こんにちは안녕하세요+*\"'()@"
		expected := "Hello世界こんにちは안녕하세요"
		result := FormatQuery(input)
		assert.Equal(t, expected, result, "Mixed character sets should be properly handled")
	})
}
