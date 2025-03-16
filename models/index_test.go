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

func TestIndexModelGet(t *testing.T) {
	// Create a test DB mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")

	// Setup config for pagination
	config.Settings.Limits.PostsPerPage = 10

	// Test case 1: Valid request with multiple threads and posts
	t.Run("Valid request with multiple threads and posts", func(t *testing.T) {
		// Mock count query for imageboards and threads
		countRows := sqlmock.NewRows([]string{"imageboards", "threads"}).AddRow(3, 25)
		mock.ExpectQuery(`SELECT \(SELECT COUNT\(\*\) FROM imageboards\) AS imageboards, \(SELECT COUNT\(\*\) FROM threads.+\) AS threads`).
			WithArgs(1).
			WillReturnRows(countRows)

		// Mock threads query
		threadRows := sqlmock.NewRows([]string{
			"thread_id", "thread_title", "thread_closed", "thread_sticky", "posts", "images",
		}).
			AddRow(1, "Thread 1", false, true, 15, 5).
			AddRow(2, "Thread 2", true, false, 10, 3)

		mock.ExpectQuery(`SELECT thread_id, thread_title, thread_closed, thread_sticky, posts, images FROM.+`).
			WithArgs(1, 0, 2).
			WillReturnRows(threadRows)

		// Mock prepare for posts query - use a more flexible regex pattern
		mock.ExpectPrepare(`SELECT \* FROM \(.+\) AS p`)

		// Mock post queries for Thread 1
		postTime := time.Now()
		thread1PostRows := sqlmock.NewRows([]string{
			"post_id", "post_num", "user_name", "user_id", "role", "post_time", "post_text",
			"image_id", "image_file", "image_thumbnail", "image_tn_height", "image_tn_width",
		}).
			AddRow(101, 1, "User1", 201, 2, postTime, "Post 1 text", 301, "image1.jpg", "thumb1.jpg", 100, 150).
			AddRow(102, 2, "User2", 202, 3, postTime, "Post 2 text", nil, nil, nil, nil, nil)

		mock.ExpectQuery(`SELECT \* FROM \(.+\) AS p`).
			WithArgs(1, 1, 3).
			WillReturnRows(thread1PostRows)

		// Mock post queries for Thread 2
		thread2PostRows := sqlmock.NewRows([]string{
			"post_id", "post_num", "user_name", "user_id", "role", "post_time", "post_text",
			"image_id", "image_file", "image_thumbnail", "image_tn_height", "image_tn_width",
		}).
			AddRow(201, 1, "User3", 203, 1, postTime, "Post 3 text", 401, "image3.jpg", "thumb3.jpg", 120, 180)

		mock.ExpectQuery(`SELECT \* FROM \(.+\) AS p`).
			WithArgs(1, 2, 3).
			WillReturnRows(thread2PostRows)

		// Create model and call Get
		model := IndexModel{
			Ib:      1,
			Page:    1,
			Threads: 2,
			Posts:   3,
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request")

		// Validate response structure
		assert.Equal(t, uint(25), model.Result.Body.Total, "Total threads should be 25")
		assert.Equal(t, uint(0), model.Result.Body.Limit, "Limit should be 0")
		assert.Equal(t, uint(2), model.Result.Body.PerPage, "Threads per page should be 2")
		assert.Equal(t, uint(13), model.Result.Body.Pages, "Total pages should be 13")
		assert.Equal(t, uint(1), model.Result.Body.CurrentPage, "Current page should be 1")

		// Check threads
		threads, ok := model.Result.Body.Items.([]IndexThreadHeader)
		assert.True(t, ok, "Items should be of type []IndexThreadHeader")
		assert.Equal(t, 2, len(threads), "Should have 2 threads")

		// Check first thread
		assert.Equal(t, uint(1), threads[0].ID, "First thread ID should be 1")
		assert.Equal(t, "Thread 1", threads[0].Title, "First thread title should match")
		assert.Equal(t, false, threads[0].Closed, "First thread closed status should match")
		assert.Equal(t, true, threads[0].Sticky, "First thread sticky status should match")
		assert.Equal(t, uint(15), threads[0].Total, "First thread post count should match")
		assert.Equal(t, uint(5), threads[0].Images, "First thread image count should match")
		assert.Equal(t, uint(2), threads[0].Pages, "First thread pages should be 2")
		assert.Equal(t, 2, len(threads[0].Posts), "First thread should have 2 posts")

		// Check first thread's first post
		post1 := threads[0].Posts[0]
		assert.Equal(t, uint(101), post1.ID, "First post ID should be 101")
		assert.Equal(t, uint(1), post1.Num, "First post number should be 1")
		assert.Equal(t, "User1", post1.Name, "First post username should match")
		assert.Equal(t, uint(201), post1.UID, "First post user ID should match")
		assert.Equal(t, uint(2), post1.Group, "First post role should match")
		assert.Equal(t, postTime.Unix(), post1.Time.Unix(), "First post time should match")
		assert.Equal(t, "Post 1 text", *post1.Text, "First post text should match")
		assert.NotNil(t, post1.ImageID, "First post image ID should not be nil")
		assert.Equal(t, uint(301), *post1.ImageID, "First post image ID should match")
		assert.Equal(t, "image1.jpg", *post1.File, "First post filename should match")
		assert.Equal(t, "thumb1.jpg", *post1.Thumb, "First post thumbnail should match")
		assert.Equal(t, uint(100), *post1.ThumbHeight, "First post thumbnail height should match")
		assert.Equal(t, uint(150), *post1.ThumbWidth, "First post thumbnail width should match")

		// Check first thread's second post (no image)
		post2 := threads[0].Posts[1]
		assert.Equal(t, uint(102), post2.ID, "Second post ID should be 102")
		assert.Equal(t, "Post 2 text", *post2.Text, "Second post text should match")
		assert.Nil(t, post2.ImageID, "Second post image ID should be nil")
		assert.Nil(t, post2.File, "Second post filename should be nil")

		// Check second thread
		assert.Equal(t, uint(2), threads[1].ID, "Second thread ID should be 2")
		assert.Equal(t, "Thread 2", threads[1].Title, "Second thread title should match")
		assert.Equal(t, true, threads[1].Closed, "Second thread closed status should match")
		assert.Equal(t, false, threads[1].Sticky, "Second thread sticky status should match")
		assert.Equal(t, uint(10), threads[1].Total, "Second thread post count should match")
		assert.Equal(t, uint(3), threads[1].Images, "Second thread image count should match")
		assert.Equal(t, uint(1), threads[1].Pages, "Second thread pages should be 1")
		assert.Equal(t, 1, len(threads[1].Posts), "Second thread should have 1 post")
	})

	// Test case 2: Empty parameters
	t.Run("Empty parameters", func(t *testing.T) {
		model := IndexModel{
			Ib:      0,
			Page:    0,
			Threads: 2,
			Posts:   3,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for empty parameters")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 3: Missing imageboard ID
	t.Run("Missing imageboard ID", func(t *testing.T) {
		model := IndexModel{
			Ib:      0, // Missing imageboard ID
			Page:    1,
			Threads: 2,
			Posts:   3,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for missing imageboard ID")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 4: Missing page
	t.Run("Missing page", func(t *testing.T) {
		model := IndexModel{
			Ib:      1,
			Page:    0, // Missing page
			Threads: 2,
			Posts:   3,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for missing page")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 5: Page exceeds total pages
	t.Run("Page exceeds total pages", func(t *testing.T) {
		// Mock count query
		countRows := sqlmock.NewRows([]string{"imageboards", "threads"}).AddRow(3, 25)
		mock.ExpectQuery(`SELECT \(SELECT COUNT\(\*\) FROM imageboards\) AS imageboards, \(SELECT COUNT\(\*\) FROM threads.+\) AS threads`).
			WithArgs(1).
			WillReturnRows(countRows)

		model := IndexModel{
			Ib:      1,
			Page:    20, // Exceeds total pages (should be 13 with 25 total threads and 2 per page)
			Threads: 2,
			Posts:   3,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for page exceeding total")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 6: Invalid imageboard ID
	t.Run("Invalid imageboard ID", func(t *testing.T) {
		// Mock count query with 2 imageboards
		countRows := sqlmock.NewRows([]string{"imageboards", "threads"}).AddRow(2, 25)
		mock.ExpectQuery(`SELECT \(SELECT COUNT\(\*\) FROM imageboards\) AS imageboards, \(SELECT COUNT\(\*\) FROM threads.+\) AS threads`).
			WithArgs(3). // Imageboard ID 3 doesn't exist
			WillReturnRows(countRows)

		model := IndexModel{
			Ib:      3, // Exceeds total imageboards (2)
			Page:    1,
			Threads: 2,
			Posts:   3,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for invalid imageboard ID")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 7: Database connection error
	t.Run("Database connection error", func(t *testing.T) {
		// Force error by closing the mock db
		db.CloseDb()

		model := IndexModel{
			Ib:      1,
			Page:    1,
			Threads: 2,
			Posts:   3,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for DB connection failure")
	})

	// Test case 8: Error in count query
	t.Run("Error in count query", func(t *testing.T) {
		// Restore db connection
		mock, err = db.NewTestDb()
		assert.NoError(t, err, "An error was not expected")

		mock.ExpectQuery(`SELECT \(SELECT COUNT\(\*\) FROM imageboards\) AS imageboards, \(SELECT COUNT\(\*\) FROM threads.+\) AS threads`).
			WithArgs(1).
			WillReturnError(sqlmock.ErrCancelled)

		model := IndexModel{
			Ib:      1,
			Page:    1,
			Threads: 2,
			Posts:   3,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for count query failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 9: Error in threads query
	t.Run("Error in threads query", func(t *testing.T) {
		// Mock count query
		countRows := sqlmock.NewRows([]string{"imageboards", "threads"}).AddRow(3, 25)
		mock.ExpectQuery(`SELECT \(SELECT COUNT\(\*\) FROM imageboards\) AS imageboards, \(SELECT COUNT\(\*\) FROM threads.+\) AS threads`).
			WithArgs(1).
			WillReturnRows(countRows)

		// Error in threads query
		mock.ExpectQuery(`SELECT thread_id, thread_title, thread_closed, thread_sticky, posts, images FROM.+`).
			WithArgs(1, 0, 2).
			WillReturnError(sqlmock.ErrCancelled)

		model := IndexModel{
			Ib:      1,
			Page:    1,
			Threads: 2,
			Posts:   3,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for threads query failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 10: Error in scan (threads)
	t.Run("Error in scan (threads)", func(t *testing.T) {
		// Mock count query
		countRows := sqlmock.NewRows([]string{"imageboards", "threads"}).AddRow(3, 25)
		mock.ExpectQuery(`SELECT \(SELECT COUNT\(\*\) FROM imageboards\) AS imageboards, \(SELECT COUNT\(\*\) FROM threads.+\) AS threads`).
			WithArgs(1).
			WillReturnRows(countRows)

		// Create row with wrong number of columns to cause scan error
		threadRows := sqlmock.NewRows([]string{
			"thread_id", "thread_title", // Missing columns
		}).AddRow(1, "Thread 1")

		mock.ExpectQuery(`SELECT thread_id, thread_title, thread_closed, thread_sticky, posts, images FROM.+`).
			WithArgs(1, 0, 2).
			WillReturnRows(threadRows)

		model := IndexModel{
			Ib:      1,
			Page:    1,
			Threads: 2,
			Posts:   3,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for thread scan failure")
	})

	// Test case 11: Error in prepare
	t.Run("Error in prepare", func(t *testing.T) {
		// Mock count query
		countRows := sqlmock.NewRows([]string{"imageboards", "threads"}).AddRow(3, 25)
		mock.ExpectQuery(`SELECT \(SELECT COUNT\(\*\) FROM imageboards\) AS imageboards, \(SELECT COUNT\(\*\) FROM threads.+\) AS threads`).
			WithArgs(1).
			WillReturnRows(countRows)

		// Mock threads query
		threadRows := sqlmock.NewRows([]string{
			"thread_id", "thread_title", "thread_closed", "thread_sticky", "posts", "images",
		}).
			AddRow(1, "Thread 1", false, true, 15, 5)

		mock.ExpectQuery(`SELECT thread_id, thread_title, thread_closed, thread_sticky, posts, images FROM.+`).
			WithArgs(1, 0, 2).
			WillReturnRows(threadRows)

		// Error in prepare - use a more flexible regex pattern
		mock.ExpectPrepare(`SELECT \* FROM \(.+\) AS p`).
			WillReturnError(sqlmock.ErrCancelled)

		model := IndexModel{
			Ib:      1,
			Page:    1,
			Threads: 2,
			Posts:   3,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for prepare failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 12: Error in posts query
	t.Run("Error in posts query", func(t *testing.T) {
		// Mock count query
		countRows := sqlmock.NewRows([]string{"imageboards", "threads"}).AddRow(3, 25)
		mock.ExpectQuery(`SELECT \(SELECT COUNT\(\*\) FROM imageboards\) AS imageboards, \(SELECT COUNT\(\*\) FROM threads.+\) AS threads`).
			WithArgs(1).
			WillReturnRows(countRows)

		// Mock threads query
		threadRows := sqlmock.NewRows([]string{
			"thread_id", "thread_title", "thread_closed", "thread_sticky", "posts", "images",
		}).
			AddRow(1, "Thread 1", false, true, 15, 5)

		mock.ExpectQuery(`SELECT thread_id, thread_title, thread_closed, thread_sticky, posts, images FROM.+`).
			WithArgs(1, 0, 2).
			WillReturnRows(threadRows)

		// Mock prepare - use a more flexible regex pattern
		mock.ExpectPrepare(`SELECT \* FROM \(.+\) AS p`)

		// Error in post query - use a more flexible regex pattern
		mock.ExpectQuery(`SELECT \* FROM \(.+\) AS p`).
			WithArgs(1, 1, 3).
			WillReturnError(sqlmock.ErrCancelled)

		model := IndexModel{
			Ib:      1,
			Page:    1,
			Threads: 2,
			Posts:   3,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for posts query failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 13: Error in post scan
	t.Run("Error in post scan", func(t *testing.T) {
		// Mock count query
		countRows := sqlmock.NewRows([]string{"imageboards", "threads"}).AddRow(3, 25)
		mock.ExpectQuery(`SELECT \(SELECT COUNT\(\*\) FROM imageboards\) AS imageboards, \(SELECT COUNT\(\*\) FROM threads.+\) AS threads`).
			WithArgs(1).
			WillReturnRows(countRows)

		// Mock threads query
		threadRows := sqlmock.NewRows([]string{
			"thread_id", "thread_title", "thread_closed", "thread_sticky", "posts", "images",
		}).
			AddRow(1, "Thread 1", false, true, 15, 5)

		mock.ExpectQuery(`SELECT thread_id, thread_title, thread_closed, thread_sticky, posts, images FROM.+`).
			WithArgs(1, 0, 2).
			WillReturnRows(threadRows)

		// Mock prepare - use a more flexible regex pattern
		mock.ExpectPrepare(`SELECT \* FROM \(.+\) AS p`)

		// Create post row with wrong number of columns to cause scan error
		postRows := sqlmock.NewRows([]string{
			"post_id", "post_num", // Missing columns
		}).AddRow(101, 1)

		// Use a more flexible regex pattern
		mock.ExpectQuery(`SELECT \* FROM \(.+\) AS p`).
			WithArgs(1, 1, 3).
			WillReturnRows(postRows)

		model := IndexModel{
			Ib:      1,
			Page:    1,
			Threads: 2,
			Posts:   3,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for post scan failure")
	})

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All mock expectations should be met")
}
