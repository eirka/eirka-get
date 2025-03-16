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

func TestPostModelGet(t *testing.T) {
	// Create a test DB mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")

	// Test case 1: Valid request with post containing an image
	t.Run("Valid request with post containing image", func(t *testing.T) {
		// Mock post query
		postTime := time.Now()
		postRows := sqlmock.NewRows([]string{
			"thread_id", "post_id", "post_num", "user_name", "user_id", "role", 
			"post_time", "post_text", "image_id", "image_file", "image_thumbnail", 
			"image_tn_height", "image_tn_width",
		}).AddRow(
			42, 101, 5, "TestUser", 123, 2, 
			postTime, "This is a test post", 789, "test.jpg", "test_thumb.jpg", 
			100, 150,
		)

		mock.ExpectQuery(`SELECT threads.thread_id, posts.post_id, post_num, user_name, users.user_id, COALESCE.+`).
			WithArgs(1, 5, 42, 1).
			WillReturnRows(postRows)

		// Create model and call Get
		model := PostModel{
			Ib:     1,
			Thread: 42,
			ID:     5,
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request")

		// Validate response structure
		post := model.Result.Body
		assert.Equal(t, uint(42), post.ThreadID, "Thread ID should be 42")
		assert.Equal(t, uint(101), post.PostID, "Post ID should be 101")
		assert.Equal(t, uint(5), post.Num, "Post number should be 5")
		assert.Equal(t, "TestUser", post.Name, "Username should match")
		assert.Equal(t, uint(123), post.UID, "User ID should match")
		assert.Equal(t, uint(2), post.Group, "User role should match")
		assert.Equal(t, postTime.Unix(), post.Time.Unix(), "Post time should match")
		assert.Equal(t, "This is a test post", *post.Text, "Post text should match")
		assert.NotNil(t, post.ImageID, "Image ID should not be nil")
		assert.Equal(t, uint(789), *post.ImageID, "Image ID should match")
		assert.Equal(t, "test.jpg", *post.File, "Image filename should match")
		assert.Equal(t, "test_thumb.jpg", *post.Thumb, "Thumbnail should match")
		assert.Equal(t, uint(100), *post.ThumbHeight, "Thumbnail height should match")
		assert.Equal(t, uint(150), *post.ThumbWidth, "Thumbnail width should match")
	})

	// Test case 2: Valid request with post without an image
	t.Run("Valid request with post without image", func(t *testing.T) {
		// Mock post query
		postTime := time.Now()
		postText := "This is a test post without image"
		postRows := sqlmock.NewRows([]string{
			"thread_id", "post_id", "post_num", "user_name", "user_id", "role", 
			"post_time", "post_text", "image_id", "image_file", "image_thumbnail", 
			"image_tn_height", "image_tn_width",
		}).AddRow(
			42, 101, 5, "TestUser", 123, 2, 
			postTime, postText, nil, nil, nil, 
			nil, nil,
		)

		mock.ExpectQuery(`SELECT threads.thread_id, posts.post_id, post_num, user_name, users.user_id, COALESCE.+`).
			WithArgs(1, 5, 42, 1).
			WillReturnRows(postRows)

		// Create model and call Get
		model := PostModel{
			Ib:     1,
			Thread: 42,
			ID:     5,
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request")

		// Validate response structure
		post := model.Result.Body
		assert.Equal(t, uint(42), post.ThreadID, "Thread ID should be 42")
		assert.Equal(t, uint(101), post.PostID, "Post ID should be 101")
		assert.Equal(t, uint(5), post.Num, "Post number should be 5")
		assert.Equal(t, "TestUser", post.Name, "Username should match")
		assert.Equal(t, uint(123), post.UID, "User ID should match")
		assert.Equal(t, uint(2), post.Group, "User role should match")
		assert.Equal(t, postTime.Unix(), post.Time.Unix(), "Post time should match")
		assert.Equal(t, postText, *post.Text, "Post text should match")
		assert.Nil(t, post.ImageID, "Image ID should be nil")
		assert.Nil(t, post.File, "Image filename should be nil")
		assert.Nil(t, post.Thumb, "Thumbnail should be nil")
		assert.Nil(t, post.ThumbHeight, "Thumbnail height should be nil")
		assert.Nil(t, post.ThumbWidth, "Thumbnail width should be nil")
	})

	// Test case 3: Empty parameters
	t.Run("Empty parameters", func(t *testing.T) {
		model := PostModel{
			Ib:     0,
			Thread: 0,
			ID:     0,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for empty parameters")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 4: Missing imageboard ID
	t.Run("Missing imageboard ID", func(t *testing.T) {
		model := PostModel{
			Ib:     0, // Missing imageboard ID
			Thread: 42,
			ID:     5,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for missing imageboard ID")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 5: Missing thread ID
	t.Run("Missing thread ID", func(t *testing.T) {
		model := PostModel{
			Ib:     1,
			Thread: 0, // Missing thread ID
			ID:     5,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for missing thread ID")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 6: Missing post number
	t.Run("Missing post number", func(t *testing.T) {
		model := PostModel{
			Ib:     1,
			Thread: 42,
			ID:     0, // Missing post number
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for missing post number")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 7: Post not found
	t.Run("Post not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT threads.thread_id, posts.post_id, post_num, user_name, users.user_id, COALESCE.+`).
			WithArgs(1, 5, 42, 1).
			WillReturnError(sql.ErrNoRows)

		model := PostModel{
			Ib:     1,
			Thread: 42,
			ID:     5,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for post not found")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 8: Database connection error
	t.Run("Database connection error", func(t *testing.T) {
		// Force error by closing the mock db
		db.CloseDb()

		model := PostModel{
			Ib:     1,
			Thread: 42,
			ID:     5,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for DB connection failure")
	})

	// Test case 9: Query execution error
	t.Run("Query execution error", func(t *testing.T) {
		// Restore db connection
		mock, err = db.NewTestDb()
		assert.NoError(t, err, "An error was not expected")

		mock.ExpectQuery(`SELECT threads.thread_id, posts.post_id, post_num, user_name, users.user_id, COALESCE.+`).
			WithArgs(1, 5, 42, 1).
			WillReturnError(sqlmock.ErrCancelled)

		model := PostModel{
			Ib:     1,
			Thread: 42,
			ID:     5,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for query execution failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 10: Scan error
	t.Run("Scan error", func(t *testing.T) {
		// Create row with wrong number of columns to cause scan error
		postRows := sqlmock.NewRows([]string{
			"thread_id", "post_id", // Missing most columns
		}).AddRow(42, 101)

		mock.ExpectQuery(`SELECT threads.thread_id, posts.post_id, post_num, user_name, users.user_id, COALESCE.+`).
			WithArgs(1, 5, 42, 1).
			WillReturnRows(postRows)

		model := PostModel{
			Ib:     1,
			Thread: 42,
			ID:     5,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for scan failure")
	})

	// Test case 11: Post with null values for optional fields
	t.Run("Post with null values for optional fields", func(t *testing.T) {
		// Mock post query with null values for text and image fields
		postTime := time.Now()
		postRows := sqlmock.NewRows([]string{
			"thread_id", "post_id", "post_num", "user_name", "user_id", "role", 
			"post_time", "post_text", "image_id", "image_file", "image_thumbnail", 
			"image_tn_height", "image_tn_width",
		}).AddRow(
			42, 101, 5, "TestUser", 123, 2, 
			postTime, nil, nil, nil, nil, 
			nil, nil,
		)

		mock.ExpectQuery(`SELECT threads.thread_id, posts.post_id, post_num, user_name, users.user_id, COALESCE.+`).
			WithArgs(1, 5, 42, 1).
			WillReturnRows(postRows)

		// Create model and call Get
		model := PostModel{
			Ib:     1,
			Thread: 42,
			ID:     5,
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request")

		// Validate response structure
		post := model.Result.Body
		assert.Equal(t, uint(42), post.ThreadID, "Thread ID should be 42")
		assert.Equal(t, uint(101), post.PostID, "Post ID should be 101")
		assert.Nil(t, post.Text, "Post text should be nil")
		assert.Nil(t, post.ImageID, "Image ID should be nil")
	})

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All mock expectations should be met")
}