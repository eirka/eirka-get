package models

import (
	"testing"

	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	
	u "github.com/eirka/eirka-get/utils"
)

func TestTagSearchModelGet(t *testing.T) {
	// Create a test DB mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")

	// Setup config for validation
	config.Settings.Limits.TagMaxLength = 30
	config.Settings.Limits.TagMinLength = 3

	// Test case 1: Valid search with multiple results
	t.Run("Valid search with multiple results", func(t *testing.T) {
		// Mock tag search query
		tagRows := sqlmock.NewRows([]string{
			"count", "tag_id", "tag_name", "tagtype_id",
		}).
			AddRow(50, 1, "anime character", 1).
			AddRow(30, 2, "anime series", 2).
			AddRow(20, 3, "anime artist", 3)

		mock.ExpectQuery(`SELECT count, tag_id, tag_name, tagtype_id FROM \(.+\) AS search`).
			WithArgs("anime", "anime", "anime", 1).
			WillReturnRows(tagRows)

		// Create model and call Get
		model := TagSearchModel{
			Ib:   1,
			Term: "anime",
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid search")

		// Validate response structure
		tags := model.Result.Body
		assert.Equal(t, 3, len(tags), "Should have 3 tags")

		// Check first tag
		assert.Equal(t, uint(1), tags[0].ID, "First tag ID should be 1")
		assert.Equal(t, "anime character", tags[0].Tag, "First tag name should match")
		assert.Equal(t, uint(50), tags[0].Total, "First tag count should match")
		assert.Equal(t, uint(1), tags[0].Type, "First tag type should match")

		// Check second tag
		assert.Equal(t, uint(2), tags[1].ID, "Second tag ID should be 2")
		assert.Equal(t, "anime series", tags[1].Tag, "Second tag name should match")
		assert.Equal(t, uint(30), tags[1].Total, "Second tag count should match")
		assert.Equal(t, uint(2), tags[1].Type, "Second tag type should match")

		// Check third tag
		assert.Equal(t, uint(3), tags[2].ID, "Third tag ID should be 3")
		assert.Equal(t, "anime artist", tags[2].Tag, "Third tag name should match")
		assert.Equal(t, uint(20), tags[2].Total, "Third tag count should match")
		assert.Equal(t, uint(3), tags[2].Type, "Third tag type should match")
	})

	// Test case 2: Valid search with multiple words
	t.Run("Valid search with multiple words", func(t *testing.T) {
		// Mock tag search query
		tagRows := sqlmock.NewRows([]string{
			"count", "tag_id", "tag_name", "tagtype_id",
		}).
			AddRow(25, 4, "popular anime character", 1)

		mock.ExpectQuery(`SELECT count, tag_id, tag_name, tagtype_id FROM \(.+\) AS search`).
			WithArgs("popular anime", "popular", "popular", "anime", "anime", 1).
			WillReturnRows(tagRows)

		// Create model and call Get
		model := TagSearchModel{
			Ib:   1,
			Term: "popular anime",
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid search with multiple words")

		// Validate response structure
		tags := model.Result.Body
		assert.Equal(t, 1, len(tags), "Should have 1 tag")
		assert.Equal(t, "popular anime character", tags[0].Tag, "Tag name should match")
	})

	// Test case 3: Valid search with no results
	t.Run("Valid search with no results", func(t *testing.T) {
		// Mock tag search query with empty result
		tagRows := sqlmock.NewRows([]string{
			"count", "tag_id", "tag_name", "tagtype_id",
		})

		mock.ExpectQuery(`SELECT count, tag_id, tag_name, tagtype_id FROM \(.+\) AS search`).
			WithArgs("nonexistent", "nonexistent", "nonexistent", 1).
			WillReturnRows(tagRows)

		// Create model and call Get
		model := TagSearchModel{
			Ib:   1,
			Term: "nonexistent",
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid search with no results")

		// Validate response structure
		tags := model.Result.Body
		assert.Equal(t, 0, len(tags), "Should have 0 tags")
	})

	// Test case 4: Empty search term
	t.Run("Empty search term", func(t *testing.T) {
		model := TagSearchModel{
			Ib:   1,
			Term: "",
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for empty search term")
		assert.Equal(t, e.ErrNoTagName, err, "Should return ErrNoTagName")
	})

	// Test case 5: Search term too short
	t.Run("Search term too short", func(t *testing.T) {
		model := TagSearchModel{
			Ib:   1,
			Term: "ab", // Less than TagMinLength (3)
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for search term too short")
		assert.Equal(t, e.ErrTagShort, err, "Should return ErrTagShort")
	})

	// Test case 6: Search term too long
	t.Run("Search term too long", func(t *testing.T) {
		// Create a term longer than TagMaxLength (30)
		longTerm := "This is a very long search term that exceeds the maximum allowed length for tag searches"
		model := TagSearchModel{
			Ib:   1,
			Term: longTerm,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for search term too long")
		assert.Equal(t, e.ErrTagLong, err, "Should return ErrTagLong")
	})

	// Test case 7: Database connection error
	t.Run("Database connection error", func(t *testing.T) {
		// Force error by closing the mock db
		db.CloseDb()

		model := TagSearchModel{
			Ib:   1,
			Term: "anime",
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for DB connection failure")
	})

	// Test case 8: Error in search query
	t.Run("Error in search query", func(t *testing.T) {
		// Restore db connection
		mock, err = db.NewTestDb()
		assert.NoError(t, err, "An error was not expected")

		mock.ExpectQuery(`SELECT count, tag_id, tag_name, tagtype_id FROM \(.+\) AS search`).
			WithArgs("error", "error", "error", 1).
			WillReturnError(sqlmock.ErrCancelled)

		model := TagSearchModel{
			Ib:   1,
			Term: "error",
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for search query failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 9: Error in scan
	t.Run("Error in scan", func(t *testing.T) {
		// Create row with wrong number of columns to cause scan error
		tagRows := sqlmock.NewRows([]string{
			"count", "tag_id", // Missing columns
		}).AddRow(50, 1)

		mock.ExpectQuery(`SELECT count, tag_id, tag_name, tagtype_id FROM \(.+\) AS search`).
			WithArgs("scan", "scan", "scan", 1).
			WillReturnRows(tagRows)

		model := TagSearchModel{
			Ib:   1,
			Term: "scan",
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for scan failure")
	})

	// Test case 10: Search with special characters
	t.Run("Search with special characters", func(t *testing.T) {
		// Mock tag search query
		tagRows := sqlmock.NewRows([]string{
			"count", "tag_id", "tag_name", "tagtype_id",
		}).
			AddRow(10, 5, "special characters", 2)

		// The special characters should be stripped from the search term
		mock.ExpectQuery(`SELECT count, tag_id, tag_name, tagtype_id FROM \(.+\) AS search`).
			WithArgs("special@+-> characters'\"()~*", "special", "special", "characters", "characters", 1).
			WillReturnRows(tagRows)

		// Create model with special characters that should be filtered out
		model := TagSearchModel{
			Ib:   1,
			Term: "special@+-> characters'\"()~*",
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid search with special characters")

		// Validate response structure
		tags := model.Result.Body
		assert.Equal(t, 1, len(tags), "Should have 1 tag")
		assert.Equal(t, "special characters", tags[0].Tag, "Tag name should match")
	})

	// Test case 11: Test FormatQuery function directly
	t.Run("Test FormatQuery function", func(t *testing.T) {
		// Test with various special characters
		input := "test@char+act-er\"s'~*()><"
		expected := "testcharacters"
		result := u.FormatQuery(input)
		assert.Equal(t, expected, result, "Special characters should be removed")

		// Test with a clean string
		input = "cleanstring"
		expected = "cleanstring"
		result = u.FormatQuery(input)
		assert.Equal(t, expected, result, "Clean string should remain unchanged")
	})

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All mock expectations should be met")
}