package models

import (
	"testing"

	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestTagsModelGet(t *testing.T) {
	// Create a test DB mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")

	// Setup config for pagination
	config.Settings.Limits.PostsPerPage = 10

	// Test case 1: Valid request with multiple tags
	t.Run("Valid request with tags", func(t *testing.T) {
		// Mock tag count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(25)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM tags WHERE ib_id = \?`).
			WithArgs(1).
			WillReturnRows(countRows)

		// Mock tag data query
		tagRows := sqlmock.NewRows([]string{"count", "tag_id", "tag_name", "tagtype_id"}).
			AddRow(100, 1, "character", 1).
			AddRow(50, 2, "series", 2).
			AddRow(25, 3, "artist", 3)

		mock.ExpectQuery(`SELECT IFNULL\(tag_counts.count, 0\) AS count, t.tag_id, t.tag_name, t.tagtype_id FROM tags t LEFT JOIN .* tag_counts ON t.tag_id = tag_counts.tag_id WHERE t.ib_id = \? ORDER BY count DESC, t.tag_id ASC LIMIT \?, \?`).
			WithArgs(1, 0, 10).
			WillReturnRows(tagRows)

		// Create model and call Get
		model := TagsModel{
			Ib:   1,
			Page: 1,
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request")

		// Validate response structure
		assert.Equal(t, uint(25), model.Result.Body.Total, "Total should be 25")
		assert.Equal(t, uint(0), model.Result.Body.Limit, "Limit should be 0")
		assert.Equal(t, uint(10), model.Result.Body.PerPage, "PerPage should be 10")
		assert.Equal(t, uint(3), model.Result.Body.Pages, "Pages should be 3")
		assert.Equal(t, uint(1), model.Result.Body.CurrentPage, "CurrentPage should be 1")

		// Check items type and count
		tags, ok := model.Result.Body.Items.([]Tags)
		assert.True(t, ok, "Items should be of type []Tags")
		assert.Equal(t, 3, len(tags), "Should have 3 tags")

		// Check first tag data
		assert.Equal(t, uint(1), tags[0].ID, "First tag ID should be 1")
		assert.Equal(t, "character", tags[0].Tag, "First tag name should match")
		assert.Equal(t, uint(100), tags[0].Total, "First tag count should match")
		assert.Equal(t, uint(1), tags[0].Type, "First tag type should match")

		// Check second tag data
		assert.Equal(t, uint(2), tags[1].ID, "Second tag ID should be 2")
		assert.Equal(t, "series", tags[1].Tag, "Second tag name should match")
		assert.Equal(t, uint(50), tags[1].Total, "Second tag count should match")
		assert.Equal(t, uint(2), tags[1].Type, "Second tag type should match")
	})

	// Test case 2: Empty parameters
	t.Run("Empty parameters", func(t *testing.T) {
		model := TagsModel{
			Ib:   0,
			Page: 0,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for empty parameters")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 3: Page greater than total pages
	t.Run("Page exceeds total pages", func(t *testing.T) {
		// Mock tag count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(25)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM tags WHERE ib_id = \?`).
			WithArgs(1).
			WillReturnRows(countRows)

		model := TagsModel{
			Ib:   1,
			Page: 5, // Should exceed total pages which is 3
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for page exceeding total")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 4: No tags found but valid page
	t.Run("No tags found but valid page", func(t *testing.T) {
		// Mock tag count query - return 0 tags
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM tags WHERE ib_id = \?`).
			WithArgs(1).
			WillReturnRows(countRows)

		// For 0 total, we still need to set up the expectation for the query
		// because the model will still attempt to query for tags
		emptyTagRows := sqlmock.NewRows([]string{"count", "tag_id", "tag_name", "tagtype_id"})
		mock.ExpectQuery(`SELECT IFNULL\(tag_counts.count, 0\) AS count, t.tag_id, t.tag_name, t.tagtype_id FROM tags t LEFT JOIN .* tag_counts ON t.tag_id = tag_counts.tag_id WHERE t.ib_id = \? ORDER BY count DESC, t.tag_id ASC LIMIT \?, \?`).
			WithArgs(1, 0, 10).
			WillReturnRows(emptyTagRows)

		// For 0 total, we expect page 1 to be valid, but empty
		model := TagsModel{
			Ib:   1,
			Page: 1,
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid page with no tags")

		// Validate response structure
		assert.Equal(t, uint(0), model.Result.Body.Total, "Total should be 0")
		assert.Equal(t, uint(1), model.Result.Body.Pages, "Pages should be 1")

		// Check items type and count
		tags, ok := model.Result.Body.Items.([]Tags)
		assert.True(t, ok, "Items should be of type []Tags")
		assert.Equal(t, 0, len(tags), "Should have 0 tags")
	})

	// Test case 5: Database connection error
	t.Run("Database connection error", func(t *testing.T) {
		// Force error by closing the mock db
		db.CloseDb()

		model := TagsModel{
			Ib:   1,
			Page: 1,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for DB connection failure")
	})

	// Test case 6: Error in count query
	t.Run("Error in count query", func(t *testing.T) {
		// Restore db connection
		mock, err = db.NewTestDb()
		assert.NoError(t, err, "An error was not expected")

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM tags WHERE ib_id = \?`).
			WithArgs(1).
			WillReturnError(sqlmock.ErrCancelled)

		model := TagsModel{
			Ib:   1,
			Page: 1,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for count query failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 7: Error in tag data query
	t.Run("Error in tag data query", func(t *testing.T) {
		// Mock tag count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(25)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM tags WHERE ib_id = \?`).
			WithArgs(1).
			WillReturnRows(countRows)

		mock.ExpectQuery(`SELECT IFNULL\(tag_counts.count, 0\) AS count, t.tag_id, t.tag_name, t.tagtype_id FROM tags t LEFT JOIN .* tag_counts ON t.tag_id = tag_counts.tag_id WHERE t.ib_id = \? ORDER BY count DESC, t.tag_id ASC LIMIT \?, \?`).
			WithArgs(1, 0, 10).
			WillReturnError(sqlmock.ErrCancelled)

		model := TagsModel{
			Ib:   1,
			Page: 1,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for tag data query failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 8: Error in scan
	t.Run("Error in scan", func(t *testing.T) {
		// Mock tag count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(25)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM tags WHERE ib_id = \?`).
			WithArgs(1).
			WillReturnRows(countRows)

		// Create row with wrong number of columns to cause scan error
		tagRows := sqlmock.NewRows([]string{
			"count", "tag_id", // Missing columns
		}).AddRow(100, 1)

		mock.ExpectQuery(`SELECT IFNULL\(tag_counts.count, 0\) AS count, t.tag_id, t.tag_name, t.tagtype_id FROM tags t LEFT JOIN .* tag_counts ON t.tag_id = tag_counts.tag_id WHERE t.ib_id = \? ORDER BY count DESC, t.tag_id ASC LIMIT \?, \?`).
			WithArgs(1, 0, 10).
			WillReturnRows(tagRows)

		model := TagsModel{
			Ib:   1,
			Page: 1,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for scan failure")
	})

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All mock expectations should be met")
}
