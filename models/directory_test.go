package models_test

import (
	"testing"
	"time"

	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/eirka/eirka-get/models"
)

func TestDirectoryModelGet(t *testing.T) {
	// Create a test DB mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")

	// Setup config for pagination
	config.Settings.Limits.PostsPerPage = 10

	// Test case 1: Valid request with multiple threads
	t.Run("Valid request with threads", func(t *testing.T) {
		// Mock thread count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(15)
		mock.ExpectQuery(`SELECT COUNT\(DISTINCT threads.thread_id\) FROM threads WHERE threads.ib_id = \? AND threads.thread_deleted != 1 AND EXISTS \(.*\)`).
			WithArgs(1).
			WillReturnRows(countRows)

		// Mock thread data query
		threadTime := time.Now()
		threadRows := sqlmock.NewRows([]string{
			"thread_id", "thread_title", "thread_closed", "thread_sticky", "count", "image_count", "thread_last_post",
		}).
			AddRow(1, "Thread 1", false, false, 25, 5, threadTime).
			AddRow(2, "Thread 2", true, true, 15, 3, threadTime)

		mock.ExpectQuery(`SELECT threads.thread_id, thread_title, thread_closed, thread_sticky, COUNT\(posts.post_id\), COUNT\(image_id\),.*`).
			WithArgs(1, 0, 10).
			WillReturnRows(threadRows)

		// Create model and call Get
		model := models.DirectoryModel{
			Ib:   1,
			Page: 1,
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request")

		// Validate response structure
		assert.Equal(t, uint(15), model.Result.Body.Total, "Total should be 15")
		assert.Equal(t, uint(0), model.Result.Body.Limit, "Limit should be 0")
		assert.Equal(t, uint(10), model.Result.Body.PerPage, "PerPage should be 10")
		assert.Equal(t, uint(2), model.Result.Body.Pages, "Pages should be 2")
		assert.Equal(t, uint(1), model.Result.Body.CurrentPage, "CurrentPage should be 1")

		// Check items type and count
		threads, ok := model.Result.Body.Items.([]models.Directory)
		assert.True(t, ok, "Items should be of type []models.Directory")
		assert.Equal(t, 2, len(threads), "Should have 2 threads")

		// Check first thread data
		assert.Equal(t, uint(1), threads[0].ID, "First thread ID should be 1")
		assert.Equal(t, "Thread 1", threads[0].Title, "First thread title should match")
		assert.Equal(t, false, threads[0].Closed, "First thread closed status should match")
		assert.Equal(t, false, threads[0].Sticky, "First thread sticky status should match")
		assert.Equal(t, uint(25), threads[0].Posts, "First thread post count should match")
		assert.Equal(t, uint(5), threads[0].Images, "First thread image count should match")
		assert.Equal(t, uint(3), threads[0].Pages, "First thread pages should be 3")
		assert.Equal(t, threadTime.Unix(), threads[0].Last.Unix(), "First thread last post time should match")

		// Check second thread data
		assert.Equal(t, uint(2), threads[1].ID, "Second thread ID should be 2")
		assert.Equal(t, "Thread 2", threads[1].Title, "Second thread title should match")
		assert.Equal(t, true, threads[1].Closed, "Second thread closed status should match")
		assert.Equal(t, true, threads[1].Sticky, "Second thread sticky status should match")
		assert.Equal(t, uint(15), threads[1].Posts, "Second thread post count should match")
		assert.Equal(t, uint(3), threads[1].Images, "Second thread image count should match")
		assert.Equal(t, uint(2), threads[1].Pages, "Second thread pages should be 2")
		assert.Equal(t, threadTime.Unix(), threads[1].Last.Unix(), "Second thread last post time should match")
	})

	// Test case 2: Empty parameters
	t.Run("Empty parameters", func(t *testing.T) {
		model := models.DirectoryModel{
			Ib:   0,
			Page: 0,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for empty parameters")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 3: Page greater than total pages
	t.Run("Page exceeds total pages", func(t *testing.T) {
		// Mock thread count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(15)
		mock.ExpectQuery(`SELECT COUNT\(DISTINCT threads.thread_id\) FROM threads WHERE threads.ib_id = \? AND threads.thread_deleted != 1 AND EXISTS \(.*\)`).
			WithArgs(1).
			WillReturnRows(countRows)

		model := models.DirectoryModel{
			Ib:   1,
			Page: 5, // Should exceed total pages which is 2
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for page exceeding total")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 4: No threads found
	t.Run("No threads found", func(t *testing.T) {
		// Mock thread count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
		mock.ExpectQuery(`SELECT COUNT\(DISTINCT threads.thread_id\) FROM threads WHERE threads.ib_id = \? AND threads.thread_deleted != 1 AND EXISTS \(.*\)`).
			WithArgs(1).
			WillReturnRows(countRows)

		// Mock thread data query with empty result
		threadRows := sqlmock.NewRows([]string{
			"thread_id", "thread_title", "thread_closed", "thread_sticky", "count", "image_count", "thread_last_post",
		})

		mock.ExpectQuery(`SELECT threads.thread_id, thread_title, thread_closed, thread_sticky, COUNT\(posts.post_id\), COUNT\(image_id\),.*`).
			WithArgs(1, 0, 10).
			WillReturnRows(threadRows)

		model := models.DirectoryModel{
			Ib:   1,
			Page: 1,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for no threads")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 5: Database connection error
	t.Run("Database connection error", func(t *testing.T) {
		// Force error by closing the mock db
		db.CloseDb()

		model := models.DirectoryModel{
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

		mock.ExpectQuery(`SELECT COUNT\(DISTINCT threads.thread_id\) FROM threads WHERE threads.ib_id = \? AND threads.thread_deleted != 1 AND EXISTS \(.*\)`).
			WithArgs(1).
			WillReturnError(sqlmock.ErrCancelled)

		model := models.DirectoryModel{
			Ib:   1,
			Page: 1,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for count query failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 7: Error in thread data query
	t.Run("Error in thread data query", func(t *testing.T) {
		// Mock thread count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(15)
		mock.ExpectQuery(`SELECT COUNT\(DISTINCT threads.thread_id\) FROM threads WHERE threads.ib_id = \? AND threads.thread_deleted != 1 AND EXISTS \(.*\)`).
			WithArgs(1).
			WillReturnRows(countRows)

		mock.ExpectQuery(`SELECT threads.thread_id, thread_title, thread_closed, thread_sticky, COUNT\(posts.post_id\), COUNT\(image_id\),.*`).
			WithArgs(1, 0, 10).
			WillReturnError(sqlmock.ErrCancelled)

		model := models.DirectoryModel{
			Ib:   1,
			Page: 1,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for threads query failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 8: Error in scan
	t.Run("Error in scan", func(t *testing.T) {
		// Mock thread count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(15)
		mock.ExpectQuery(`SELECT COUNT\(DISTINCT threads.thread_id\) FROM threads WHERE threads.ib_id = \? AND threads.thread_deleted != 1 AND EXISTS \(.*\)`).
			WithArgs(1).
			WillReturnRows(countRows)

		// Create row with wrong number of columns to cause scan error
		threadRows := sqlmock.NewRows([]string{
			"thread_id", "thread_title", // Missing columns
		}).AddRow(1, "Thread 1")

		mock.ExpectQuery(`SELECT threads.thread_id, thread_title, thread_closed, thread_sticky, COUNT\(posts.post_id\), COUNT\(image_id\),.*`).
			WithArgs(1, 0, 10).
			WillReturnRows(threadRows)

		model := models.DirectoryModel{
			Ib:   1,
			Page: 1,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for scan failure")
	})

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All mock expectations should be met")
}
