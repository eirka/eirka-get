package models

import (
	"testing"
	"time"

	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestThreadSearchModelGet(t *testing.T) {
	// Create a test DB mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")

	// Setup config for validation and pagination
	config.Settings.Limits.TitleMaxLength = 40
	config.Settings.Limits.TitleMinLength = 3
	config.Settings.Limits.PostsPerPage = 10

	// Test case 1: Valid search with multiple results
	t.Run("Valid search with multiple results", func(t *testing.T) {
		// Mock thread search query
		searchTime := time.Now()
		threadRows := sqlmock.NewRows([]string{
			"thread_id", "thread_title", "thread_closed", "thread_sticky", "post_count", "image_count", "thread_last_post",
		}).
			AddRow(1, "Test Thread One", false, true, 15, 5, searchTime).
			AddRow(2, "Test Thread Two", true, false, 10, 3, searchTime.Add(-time.Hour))

		mock.ExpectQuery(`SELECT threads.thread_id, thread_title, thread_closed, thread_sticky, COUNT\(posts.post_id\), COUNT\(image_id\), .+ AS thread_last_post FROM threads.+`).
			WithArgs(1, "test", "search").
			WillReturnRows(threadRows)

		// Create model and call Get
		model := ThreadSearchModel{
			Ib:   1,
			Term: "test search",
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid search")

		// Validate response structure
		threads := model.Result.Body
		assert.Equal(t, 2, len(threads), "Should have 2 threads")

		// Check first thread
		assert.Equal(t, uint(1), threads[0].ID, "First thread ID should be 1")
		assert.Equal(t, "Test Thread One", threads[0].Title, "First thread title should match")
		assert.Equal(t, false, threads[0].Closed, "First thread closed status should match")
		assert.Equal(t, true, threads[0].Sticky, "First thread sticky status should match")
		assert.Equal(t, uint(15), threads[0].Posts, "First thread post count should match")
		assert.Equal(t, uint(5), threads[0].Images, "First thread image count should match")
		assert.Equal(t, uint(2), threads[0].Pages, "First thread pages should be 2")
		assert.Equal(t, searchTime.Unix(), threads[0].Last.Unix(), "First thread last post time should match")

		// Check second thread
		assert.Equal(t, uint(2), threads[1].ID, "Second thread ID should be 2")
		assert.Equal(t, "Test Thread Two", threads[1].Title, "Second thread title should match")
		assert.Equal(t, true, threads[1].Closed, "Second thread closed status should match")
		assert.Equal(t, false, threads[1].Sticky, "Second thread sticky status should match")
		assert.Equal(t, uint(10), threads[1].Posts, "Second thread post count should match")
		assert.Equal(t, uint(3), threads[1].Images, "Second thread image count should match")
		assert.Equal(t, uint(1), threads[1].Pages, "Second thread pages should be 1")
		assert.Equal(t, searchTime.Add(-time.Hour).Unix(), threads[1].Last.Unix(), "Second thread last post time should match")
	})

	// Test case 2: Valid search with no results
	t.Run("Valid search with no results", func(t *testing.T) {
		// Mock thread search query with empty result
		threadRows := sqlmock.NewRows([]string{
			"thread_id", "thread_title", "thread_closed", "thread_sticky", "post_count", "image_count", "thread_last_post",
		})

		mock.ExpectQuery(`SELECT threads.thread_id, thread_title, thread_closed, thread_sticky, COUNT\(posts.post_id\), COUNT\(image_id\), .+ AS thread_last_post FROM threads.+`).
			WithArgs(1, "unique", "term").
			WillReturnRows(threadRows)

		// Create model and call Get
		model := ThreadSearchModel{
			Ib:   1,
			Term: "unique term",
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid search with no results")

		// Validate response structure
		threads := model.Result.Body
		assert.Equal(t, 0, len(threads), "Should have 0 threads")
	})

	// Test case 3: Empty search term
	t.Run("Empty search term", func(t *testing.T) {
		model := ThreadSearchModel{
			Ib:   1,
			Term: "",
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for empty search term")
		assert.Equal(t, e.ErrNoTitle, err, "Should return ErrNoTitle")
	})

	// Test case 4: Search term too short
	t.Run("Search term too short", func(t *testing.T) {
		model := ThreadSearchModel{
			Ib:   1,
			Term: "ab", // Less than TitleMinLength (3)
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for search term too short")
		assert.Equal(t, e.ErrTitleShort, err, "Should return ErrTitleShort")
	})

	// Test case 5: Search term too long
	t.Run("Search term too long", func(t *testing.T) {
		// Create a term longer than TitleMaxLength (40)
		longTerm := "This is a very long search term that exceeds the maximum allowed length for thread title searches"
		model := ThreadSearchModel{
			Ib:   1,
			Term: longTerm,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for search term too long")
		assert.Equal(t, e.ErrTitleLong, err, "Should return ErrTitleLong")
	})

	// Test case 6: Database connection error
	t.Run("Database connection error", func(t *testing.T) {
		// Force error by closing the mock db
		db.CloseDb()

		model := ThreadSearchModel{
			Ib:   1,
			Term: "test search",
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for DB connection failure")
	})

	// Test case 7: Error in search query
	t.Run("Error in search query", func(t *testing.T) {
		// Restore db connection
		mock, err = db.NewTestDb()
		assert.NoError(t, err, "An error was not expected")

		mock.ExpectQuery(`SELECT threads.thread_id, thread_title, thread_closed, thread_sticky, COUNT\(posts.post_id\), COUNT\(image_id\), .+ AS thread_last_post FROM threads.+`).
			WithArgs(1, "test", "error").
			WillReturnError(sqlmock.ErrCancelled)

		model := ThreadSearchModel{
			Ib:   1,
			Term: "test error",
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for search query failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 8: Error in scan
	t.Run("Error in scan", func(t *testing.T) {
		// Create row with wrong number of columns to cause scan error
		threadRows := sqlmock.NewRows([]string{
			"thread_id", "thread_title", // Missing columns
		}).AddRow(1, "Test Thread")

		mock.ExpectQuery(`SELECT threads.thread_id, thread_title, thread_closed, thread_sticky, COUNT\(posts.post_id\), COUNT\(image_id\), .+ AS thread_last_post FROM threads.+`).
			WithArgs(1, "test", "scan").
			WillReturnRows(threadRows)

		model := ThreadSearchModel{
			Ib:   1,
			Term: "test scan",
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for scan failure")
	})

	// Test case 9: Search with special characters
	t.Run("Search with special characters", func(t *testing.T) {
		// Mock thread search query
		searchTime := time.Now()
		threadRows := sqlmock.NewRows([]string{
			"thread_id", "thread_title", "thread_closed", "thread_sticky", "post_count", "image_count", "thread_last_post",
		}).
			AddRow(1, "Special Characters Test", false, false, 5, 2, searchTime)

		// The special characters should be stripped from the search term
		mock.ExpectQuery(`SELECT threads.thread_id, thread_title, thread_closed, thread_sticky, COUNT\(posts.post_id\), COUNT\(image_id\), .+ AS thread_last_post FROM threads.+`).
			WithArgs(1, "special", "characters").
			WillReturnRows(threadRows)

		// Create model with special characters that should be filtered out
		model := ThreadSearchModel{
			Ib:   1,
			Term: "special@+-> characters'\"()~*",
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid search with special characters")

		// Validate response structure
		threads := model.Result.Body
		assert.Equal(t, 1, len(threads), "Should have 1 thread")
		assert.Equal(t, "Special Characters Test", threads[0].Title, "Thread title should match")
	})

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All mock expectations should be met")
}
