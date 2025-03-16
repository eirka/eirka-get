package models

import (
	"database/sql"
	"testing"
	"time"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestThreadModelGet(t *testing.T) {
	// Create a test DB mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")

	// Test case 1: Valid request with full data
	t.Run("Valid request with full data", func(t *testing.T) {
		// Mock thread info query
		threadTime := time.Now()
		threadRows := sqlmock.NewRows([]string{
			"thread_id", "thread_title", "thread_closed", "thread_sticky", "count",
		}).AddRow(1, "Test Thread", false, true, 15)

		mock.ExpectQuery(`SELECT threads.thread_id, thread_title, thread_closed, thread_sticky, COUNT\(posts.post_id\) FROM threads`).
			WithArgs(1, 1).
			WillReturnRows(threadRows)

		// Mock posts query
		postRows := sqlmock.NewRows([]string{
			"post_id", "post_num", "user_name", "user_id", "role", "post_time", "post_text",
			"image_id", "image_file", "image_thumbnail", "image_tn_height", "image_tn_width",
		}).
			AddRow(1, 1, "User1", 101, 3, threadTime, "Post 1 text", 1001, "image1.jpg", "thumb1.jpg", 100, 100).
			AddRow(2, 2, "User2", 102, 2, threadTime, "Post 2 text", nil, nil, nil, nil, nil).
			AddRow(3, 3, "User3", 103, 1, threadTime, "Post 3 text", 1003, "image3.jpg", "thumb3.jpg", 150, 150)

		mock.ExpectQuery(`SELECT posts.post_id, post_num, user_name, users.user_id,\s*COALESCE\(\s*\(SELECT MAX\(role_id\).+\) AS role,\s*post_time, post_text, image_id, image_file, image_thumbnail, image_tn_height, image_tn_width`).
			WithArgs(1, 1, 0, 15).
			WillReturnRows(postRows)

		// Create model and call Get
		model := ThreadModel{
			Ib:     1,
			Thread: 1,
			Page:   1,
			Posts:  15,
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request")

		// Validate response structure
		assert.Equal(t, uint(1), model.Result.Body.Items.(ThreadInfo).ID, "Thread ID should be 1")
		assert.Equal(t, "Test Thread", model.Result.Body.Items.(ThreadInfo).Title, "Thread title should match")
		assert.Equal(t, false, model.Result.Body.Items.(ThreadInfo).Closed, "Thread closed status should match")
		assert.Equal(t, true, model.Result.Body.Items.(ThreadInfo).Sticky, "Thread sticky status should match")
		assert.Equal(t, 3, len(model.Result.Body.Items.(ThreadInfo).Posts), "Should have 3 posts")

		// Test first post with image
		post1 := model.Result.Body.Items.(ThreadInfo).Posts[0]
		assert.Equal(t, uint(1), post1.ID, "Post ID should be 1")
		assert.Equal(t, uint(1), post1.Num, "Post number should be 1")
		assert.Equal(t, "User1", post1.Name, "Username should match")
		assert.Equal(t, uint(101), post1.UID, "User ID should match")
		assert.Equal(t, uint(3), post1.Group, "User role should match")
		assert.Equal(t, threadTime.Unix(), post1.Time.Unix(), "Post time should match")
		assert.Equal(t, "Post 1 text", *post1.Text, "Post text should match")
		assert.NotNil(t, post1.ImageID, "Image ID should not be nil")
		assert.Equal(t, uint(1001), *post1.ImageID, "Image ID should match")
		assert.Equal(t, "image1.jpg", *post1.File, "Image filename should match")
		assert.Equal(t, "thumb1.jpg", *post1.Thumb, "Thumbnail should match")
		assert.Equal(t, uint(100), *post1.ThumbHeight, "Thumbnail height should match")
		assert.Equal(t, uint(100), *post1.ThumbWidth, "Thumbnail width should match")

		// Test second post without image
		post2 := model.Result.Body.Items.(ThreadInfo).Posts[1]
		assert.Equal(t, uint(2), post2.ID, "Post ID should be 2")
		assert.Equal(t, "User2", post2.Name, "Username should match")
		assert.Nil(t, post2.ImageID, "Image ID should be nil")
		assert.Nil(t, post2.File, "File should be nil")
		assert.Nil(t, post2.Thumb, "Thumbnail should be nil")
		assert.Nil(t, post2.ThumbHeight, "Thumbnail height should be nil")
		assert.Nil(t, post2.ThumbWidth, "Thumbnail width should be nil")

		// Check pagination
		assert.Equal(t, uint(15), model.Result.Body.Total, "Total posts should be 15")
		assert.Equal(t, uint(1), model.Result.Body.Pages, "Number of pages should be 1")
		assert.Equal(t, uint(1), model.Result.Body.CurrentPage, "Current page should be 1")
		assert.Equal(t, uint(15), model.Result.Body.PerPage, "Posts per page should be 15")
	})

	// Test case 2: Page 0 request (all posts)
	t.Run("Page 0 request (all posts)", func(t *testing.T) {
		// Mock thread info query
		threadRows := sqlmock.NewRows([]string{
			"thread_id", "thread_title", "thread_closed", "thread_sticky", "count",
		}).AddRow(1, "Test Thread", false, true, 15)

		mock.ExpectQuery(`SELECT threads.thread_id, thread_title, thread_closed, thread_sticky, COUNT\(posts.post_id\) FROM threads`).
			WithArgs(1, 1).
			WillReturnRows(threadRows)

		// Since page 0 returns all posts, the limit should be 0 and per_page should be total
		postRows := sqlmock.NewRows([]string{
			"post_id", "post_num", "user_name", "user_id", "role", "post_time", "post_text",
			"image_id", "image_file", "image_thumbnail", "image_tn_height", "image_tn_width",
		}).AddRow(1, 1, "User1", 101, 3, time.Now(), "Post text", nil, nil, nil, nil, nil)

		mock.ExpectQuery(`SELECT posts.post_id, post_num, user_name, users.user_id,\s*COALESCE\(\s*\(SELECT MAX\(role_id\).+\) AS role,\s*post_time, post_text, image_id, image_file, image_thumbnail, image_tn_height, image_tn_width`).
			WithArgs(1, 1, 0, 15).
			WillReturnRows(postRows)

		// Create model and call Get
		model := ThreadModel{
			Ib:     1,
			Thread: 1,
			Page:   0, // Page 0 should return all posts
			Posts:  15,
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for page 0 request")

		// Verify pagination settings for page 0
		assert.Equal(t, uint(15), model.Result.Body.Total, "Total posts should be 15")
		assert.Equal(t, uint(15), model.Result.Body.PerPage, "PerPage should be equal to total for page 0")
		assert.Equal(t, uint(0), model.Result.Body.Limit, "Limit should be 0 for page 0")
	})

	// Test case 3: Empty parameters
	t.Run("Empty parameters", func(t *testing.T) {
		model := ThreadModel{
			Ib:     0,
			Thread: 0,
			Page:   1,
			Posts:  15,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for empty parameters")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 4: Thread not found
	t.Run("Thread not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT threads.thread_id, thread_title, thread_closed, thread_sticky, COUNT\(posts.post_id\) FROM threads`).
			WithArgs(1, 1).
			WillReturnError(sql.ErrNoRows)

		model := ThreadModel{
			Ib:     1,
			Thread: 1,
			Page:   1,
			Posts:  15,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for thread not found")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 5: Page exceeds total pages
	t.Run("Page exceeds total pages", func(t *testing.T) {
		// Mock thread info query
		threadRows := sqlmock.NewRows([]string{
			"thread_id", "thread_title", "thread_closed", "thread_sticky", "count",
		}).AddRow(1, "Test Thread", false, true, 15)

		mock.ExpectQuery(`SELECT threads.thread_id, thread_title, thread_closed, thread_sticky, COUNT\(posts.post_id\) FROM threads`).
			WithArgs(1, 1).
			WillReturnRows(threadRows)

		// Create model with page number exceeding total pages
		model := ThreadModel{
			Ib:     1,
			Thread: 1,
			Page:   10, // With 15 posts and 15 per page, there's only 1 page
			Posts:  15,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for page exceeding total")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 6: Database connection error
	t.Run("Database connection error", func(t *testing.T) {
		// Force error by closing the mock db
		db.CloseDb()

		model := ThreadModel{
			Ib:     1,
			Thread: 1,
			Page:   1,
			Posts:  15,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for DB connection failure")
	})

	// Test case 7: Error in thread info query
	t.Run("Error in thread info query", func(t *testing.T) {
		// Restore db connection
		mock, err = db.NewTestDb()
		assert.NoError(t, err, "An error was not expected")

		mock.ExpectQuery(`SELECT threads.thread_id, thread_title, thread_closed, thread_sticky, COUNT\(posts.post_id\) FROM threads`).
			WithArgs(1, 1).
			WillReturnError(sqlmock.ErrCancelled)

		model := ThreadModel{
			Ib:     1,
			Thread: 1,
			Page:   1,
			Posts:  15,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for thread info query failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 8: Error in posts query
	t.Run("Error in posts query", func(t *testing.T) {
		// Mock thread info query
		threadRows := sqlmock.NewRows([]string{
			"thread_id", "thread_title", "thread_closed", "thread_sticky", "count",
		}).AddRow(1, "Test Thread", false, true, 15)

		mock.ExpectQuery(`SELECT threads.thread_id, thread_title, thread_closed, thread_sticky, COUNT\(posts.post_id\) FROM threads`).
			WithArgs(1, 1).
			WillReturnRows(threadRows)

		// Error in posts query
		mock.ExpectQuery(`SELECT posts.post_id, post_num, user_name, users.user_id,\s*COALESCE\(\s*\(SELECT MAX\(role_id\).+\) AS role,\s*post_time, post_text, image_id, image_file, image_thumbnail, image_tn_height, image_tn_width`).
			WithArgs(1, 1, 0, 15).
			WillReturnError(sqlmock.ErrCancelled)

		model := ThreadModel{
			Ib:     1,
			Thread: 1,
			Page:   1,
			Posts:  15,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for posts query failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 9: Error in scan
	t.Run("Error in scan", func(t *testing.T) {
		// Mock thread info query
		threadRows := sqlmock.NewRows([]string{
			"thread_id", "thread_title", "thread_closed", "thread_sticky", "count",
		}).AddRow(1, "Test Thread", false, true, 15)

		mock.ExpectQuery(`SELECT threads.thread_id, thread_title, thread_closed, thread_sticky, COUNT\(posts.post_id\) FROM threads`).
			WithArgs(1, 1).
			WillReturnRows(threadRows)

		// Create row with wrong number of columns to cause scan error
		postRows := sqlmock.NewRows([]string{
			"post_id", "post_num", // Missing most columns
		}).AddRow(1, 1)

		mock.ExpectQuery(`SELECT posts.post_id, post_num, user_name, users.user_id,\s*COALESCE\(\s*\(SELECT MAX\(role_id\).+\) AS role,\s*post_time, post_text, image_id, image_file, image_thumbnail, image_tn_height, image_tn_width`).
			WithArgs(1, 1, 0, 15).
			WillReturnRows(postRows)

		model := ThreadModel{
			Ib:     1,
			Thread: 1,
			Page:   1,
			Posts:  15,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for scan failure")
	})

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All mock expectations should be met")
}
