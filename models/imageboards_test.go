package models

import (
	"testing"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestImageboardsModelGet(t *testing.T) {
	// Create a test DB mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")

	// Test case 1: Valid request with multiple imageboards
	t.Run("Valid request with multiple imageboards", func(t *testing.T) {
		// Mock imageboards query
		boardRows := sqlmock.NewRows([]string{
			"ib_id", "ib_title", "ib_description", "ib_domain", "thread_count", "post_count", "image_count",
		}).
			AddRow(1, "Anime", "Japanese Animation", "anime.example.com", 100, 1500, 800).
			AddRow(2, "Technology", "Technology Discussion", "tech.example.com", 50, 750, 300).
			AddRow(3, "Music", "Music Discussion", "music.example.com", 75, 1200, 500)

		mock.ExpectQuery(`SELECT ib_id, ib_title, ib_description, ib_domain,\s*\(SELECT COUNT\(thread_id\).+\) AS thread_count.+`).
			WillReturnRows(boardRows)

		// Create model and call Get
		model := ImageboardsModel{}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request")

		// Validate response structure
		boards := model.Result.Body
		assert.Equal(t, 3, len(boards), "Should have 3 imageboards")

		// Check first board
		assert.Equal(t, uint(1), boards[0].ID, "First board ID should be 1")
		assert.Equal(t, "Anime", boards[0].Title, "First board title should match")
		assert.Equal(t, "Japanese Animation", boards[0].Description, "First board description should match")
		assert.Equal(t, "anime.example.com", boards[0].Domain, "First board domain should match")
		assert.Equal(t, uint(100), boards[0].Threads, "First board thread count should match")
		assert.Equal(t, uint(1500), boards[0].Posts, "First board post count should match")
		assert.Equal(t, uint(800), boards[0].Images, "First board image count should match")

		// Check second board
		assert.Equal(t, uint(2), boards[1].ID, "Second board ID should be 2")
		assert.Equal(t, "Technology", boards[1].Title, "Second board title should match")
		assert.Equal(t, "Technology Discussion", boards[1].Description, "Second board description should match")
		assert.Equal(t, "tech.example.com", boards[1].Domain, "Second board domain should match")
		assert.Equal(t, uint(50), boards[1].Threads, "Second board thread count should match")
		assert.Equal(t, uint(750), boards[1].Posts, "Second board post count should match")
		assert.Equal(t, uint(300), boards[1].Images, "Second board image count should match")
	})

	// Test case 2: Valid request with single imageboard
	t.Run("Valid request with single imageboard", func(t *testing.T) {
		// Mock imageboards query
		boardRows := sqlmock.NewRows([]string{
			"ib_id", "ib_title", "ib_description", "ib_domain", "thread_count", "post_count", "image_count",
		}).
			AddRow(1, "Anime", "Japanese Animation", "anime.example.com", 100, 1500, 800)

		mock.ExpectQuery(`SELECT ib_id, ib_title, ib_description, ib_domain,\s*\(SELECT COUNT\(thread_id\).+\) AS thread_count.+`).
			WillReturnRows(boardRows)

		// Create model and call Get
		model := ImageboardsModel{}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request")

		// Validate response structure
		boards := model.Result.Body
		assert.Equal(t, 1, len(boards), "Should have 1 imageboard")
	})

	// Test case 3: No imageboards found
	t.Run("No imageboards found", func(t *testing.T) {
		// Mock empty imageboards query
		boardRows := sqlmock.NewRows([]string{
			"ib_id", "ib_title", "ib_description", "ib_domain", "thread_count", "post_count", "image_count",
		})

		mock.ExpectQuery(`SELECT ib_id, ib_title, ib_description, ib_domain,\s*\(SELECT COUNT\(thread_id\).+\) AS thread_count.+`).
			WillReturnRows(boardRows)

		// Create model and call Get
		model := ImageboardsModel{}

		err := model.Get()
		assert.Error(t, err, "Should return error for no imageboards")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 4: Database connection error
	t.Run("Database connection error", func(t *testing.T) {
		// Force error by closing the mock db
		db.CloseDb()

		model := ImageboardsModel{}

		err := model.Get()
		assert.Error(t, err, "Should return error for DB connection failure")
	})

	// Test case 5: Error in query execution
	t.Run("Error in query execution", func(t *testing.T) {
		// Restore db connection
		mock, err = db.NewTestDb()
		assert.NoError(t, err, "An error was not expected")

		mock.ExpectQuery(`SELECT ib_id, ib_title, ib_description, ib_domain,\s*\(SELECT COUNT\(thread_id\).+\) AS thread_count.+`).
			WillReturnError(sqlmock.ErrCancelled)

		model := ImageboardsModel{}

		err = model.Get()
		assert.Error(t, err, "Should return error for query execution failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 6: Error in scan
	t.Run("Error in scan", func(t *testing.T) {
		// Create row with wrong number of columns to cause scan error
		boardRows := sqlmock.NewRows([]string{
			"ib_id", "ib_title", "ib_description", // Missing columns
		}).AddRow(1, "Anime", "Japanese Animation")

		mock.ExpectQuery(`SELECT ib_id, ib_title, ib_description, ib_domain,\s*\(SELECT COUNT\(thread_id\).+\) AS thread_count.+`).
			WillReturnRows(boardRows)

		model := ImageboardsModel{}

		err = model.Get()
		assert.Error(t, err, "Should return error for scan failure")
	})

	// Test case 7: Zero counts
	t.Run("Zero counts", func(t *testing.T) {
		// Mock imageboards query with zero counts
		boardRows := sqlmock.NewRows([]string{
			"ib_id", "ib_title", "ib_description", "ib_domain", "thread_count", "post_count", "image_count",
		}).
			AddRow(1, "Empty Board", "No content yet", "empty.example.com", 0, 0, 0)

		mock.ExpectQuery(`SELECT ib_id, ib_title, ib_description, ib_domain,\s*\(SELECT COUNT\(thread_id\).+\) AS thread_count.+`).
			WillReturnRows(boardRows)

		// Create model and call Get
		model := ImageboardsModel{}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request with zero counts")

		// Validate response structure
		boards := model.Result.Body
		assert.Equal(t, 1, len(boards), "Should have 1 imageboard")
		assert.Equal(t, "Empty Board", boards[0].Title, "Board title should match")
		assert.Equal(t, uint(0), boards[0].Threads, "Thread count should be 0")
		assert.Equal(t, uint(0), boards[0].Posts, "Post count should be 0")
		assert.Equal(t, uint(0), boards[0].Images, "Image count should be 0")
	})

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All mock expectations should be met")
}