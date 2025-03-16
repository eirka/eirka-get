package models

import (
	"database/sql"
	"testing"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestImageModelGet(t *testing.T) {
	// Create a test DB mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")

	// Test case 1: Valid request with image, prev/next images, and tags
	t.Run("Valid request with full data", func(t *testing.T) {
		// Mock image info query
		imageRows := sqlmock.NewRows([]string{
			"image_id", "thread_id", "post_num", "post_id", "image_file", "image_orig_height", "image_orig_width",
		}).AddRow(5, 10, 3, 15, "test.jpg", 800, 600)

		mock.ExpectQuery(`SELECT image_id, posts.thread_id, posts.post_num, posts.post_id, image_file, image_orig_height, image_orig_width FROM images`).
			WithArgs(5, 1).
			WillReturnRows(imageRows)

		// Mock prev/next query with both prev and next
		prevNextRows := sqlmock.NewRows([]string{"previous", "next"}).
			AddRow(4, 6)

		mock.ExpectQuery(`SELECT \(SELECT image_id FROM images INNER JOIN posts.+\) AS previous, \(SELECT image_id.+\) AS next`).
			WithArgs(10, 5, 10, 5).
			WillReturnRows(prevNextRows)

		// Mock tags query
		tagRows := sqlmock.NewRows([]string{"tag_id", "tagtype_id", "tag_name"}).
			AddRow(1, "character", "test tag 1").
			AddRow(2, "copyright", "test tag 2")

		mock.ExpectQuery(`SELECT tags.tag_id, tagtype_id, tag_name FROM tagmap LEFT JOIN tags`).
			WithArgs(5).
			WillReturnRows(tagRows)

		// Create model and call Get
		model := ImageModel{
			Ib: 1,
			ID: 5,
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request")

		// Validate response structure
		assert.Equal(t, uint(5), model.Result.Body.ID, "Image ID should be 5")
		assert.Equal(t, uint(10), model.Result.Body.Thread, "Thread ID should be 10")
		assert.Equal(t, uint(3), model.Result.Body.PostNum, "Post number should be 3")
		assert.Equal(t, uint(15), model.Result.Body.PostID, "Post ID should be 15")
		assert.Equal(t, "test.jpg", model.Result.Body.File, "Filename should match")
		assert.Equal(t, uint(800), model.Result.Body.Height, "Height should match")
		assert.Equal(t, uint(600), model.Result.Body.Width, "Width should match")

		// Check prev/next
		assert.NotNil(t, model.Result.Body.Prev, "Prev should not be nil")
		assert.Equal(t, uint(4), *model.Result.Body.Prev, "Prev image ID should be 4")
		assert.NotNil(t, model.Result.Body.Next, "Next should not be nil")
		assert.Equal(t, uint(6), *model.Result.Body.Next, "Next image ID should be 6")

		// Check tags
		assert.Equal(t, 2, len(model.Result.Body.Tags), "Should have 2 tags")
		assert.Equal(t, uint(1), model.Result.Body.Tags[0].ID, "First tag ID should be 1")
		assert.Equal(t, "character", model.Result.Body.Tags[0].Type, "First tag type should be character")
		assert.Equal(t, "test tag 1", model.Result.Body.Tags[0].Tag, "First tag name should match")
		assert.Equal(t, uint(2), model.Result.Body.Tags[1].ID, "Second tag ID should be 2")
		assert.Equal(t, "copyright", model.Result.Body.Tags[1].Type, "Second tag type should be copyright")
		assert.Equal(t, "test tag 2", model.Result.Body.Tags[1].Tag, "Second tag name should match")
	})

	// Test case 2: Valid request with no prev/next images
	t.Run("Valid request with no prev/next", func(t *testing.T) {
		// Mock image info query
		imageRows := sqlmock.NewRows([]string{
			"image_id", "thread_id", "post_num", "post_id", "image_file", "image_orig_height", "image_orig_width",
		}).AddRow(5, 10, 3, 15, "test.jpg", 800, 600)

		mock.ExpectQuery(`SELECT image_id, posts.thread_id, posts.post_num, posts.post_id, image_file, image_orig_height, image_orig_width FROM images`).
			WithArgs(5, 1).
			WillReturnRows(imageRows)

		// Mock prev/next query with NULL values
		prevNextRows := sqlmock.NewRows([]string{"previous", "next"}).
			AddRow(nil, nil)

		mock.ExpectQuery(`SELECT \(SELECT image_id FROM images INNER JOIN posts.+\) AS previous, \(SELECT image_id.+\) AS next`).
			WithArgs(10, 5, 10, 5).
			WillReturnRows(prevNextRows)

		// Mock tags query with no tags
		tagRows := sqlmock.NewRows([]string{"tag_id", "tagtype_id", "tag_name"})

		mock.ExpectQuery(`SELECT tags.tag_id, tagtype_id, tag_name FROM tagmap LEFT JOIN tags`).
			WithArgs(5).
			WillReturnRows(tagRows)

		// Create model and call Get
		model := ImageModel{
			Ib: 1,
			ID: 5,
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request")

		// Validate response structure
		assert.Equal(t, uint(5), model.Result.Body.ID, "Image ID should be 5")
		assert.Nil(t, model.Result.Body.Prev, "Prev should be nil")
		assert.Nil(t, model.Result.Body.Next, "Next should be nil")
		assert.Empty(t, model.Result.Body.Tags, "Tags should be empty")
	})

	// Test case 3: Empty parameters
	t.Run("Empty parameters", func(t *testing.T) {
		model := ImageModel{
			Ib: 0,
			ID: 0,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for empty parameters")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 4: Image not found
	t.Run("Image not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT image_id, posts.thread_id, posts.post_num, posts.post_id, image_file, image_orig_height, image_orig_width FROM images`).
			WithArgs(5, 1).
			WillReturnError(sql.ErrNoRows)

		model := ImageModel{
			Ib: 1,
			ID: 5,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for image not found")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 5: Database connection error
	t.Run("Database connection error", func(t *testing.T) {
		// Force error by closing the mock db
		db.CloseDb()

		model := ImageModel{
			Ib: 1,
			ID: 5,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for DB connection failure")
	})

	// Test case 6: Error in image info query
	t.Run("Error in image info query", func(t *testing.T) {
		// Restore db connection
		mock, err = db.NewTestDb()
		assert.NoError(t, err, "An error was not expected")

		mock.ExpectQuery(`SELECT image_id, posts.thread_id, posts.post_num, posts.post_id, image_file, image_orig_height, image_orig_width FROM images`).
			WithArgs(5, 1).
			WillReturnError(sqlmock.ErrCancelled)

		model := ImageModel{
			Ib: 1,
			ID: 5,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for image info query failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 7: Error in prev/next query
	t.Run("Error in prev/next query", func(t *testing.T) {
		// Mock image info query
		imageRows := sqlmock.NewRows([]string{
			"image_id", "thread_id", "post_num", "post_id", "image_file", "image_orig_height", "image_orig_width",
		}).AddRow(5, 10, 3, 15, "test.jpg", 800, 600)

		mock.ExpectQuery(`SELECT image_id, posts.thread_id, posts.post_num, posts.post_id, image_file, image_orig_height, image_orig_width FROM images`).
			WithArgs(5, 1).
			WillReturnRows(imageRows)

		// Error in prev/next query
		mock.ExpectQuery(`SELECT \(SELECT image_id FROM images INNER JOIN posts.+\) AS previous, \(SELECT image_id.+\) AS next`).
			WithArgs(10, 5, 10, 5).
			WillReturnError(sqlmock.ErrCancelled)

		model := ImageModel{
			Ib: 1,
			ID: 5,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for prev/next query failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 8: Error in tags query
	t.Run("Error in tags query", func(t *testing.T) {
		// Mock image info query
		imageRows := sqlmock.NewRows([]string{
			"image_id", "thread_id", "post_num", "post_id", "image_file", "image_orig_height", "image_orig_width",
		}).AddRow(5, 10, 3, 15, "test.jpg", 800, 600)

		mock.ExpectQuery(`SELECT image_id, posts.thread_id, posts.post_num, posts.post_id, image_file, image_orig_height, image_orig_width FROM images`).
			WithArgs(5, 1).
			WillReturnRows(imageRows)

		// Mock prev/next query
		prevNextRows := sqlmock.NewRows([]string{"previous", "next"}).
			AddRow(4, 6)

		mock.ExpectQuery(`SELECT \(SELECT image_id FROM images INNER JOIN posts.+\) AS previous, \(SELECT image_id.+\) AS next`).
			WithArgs(10, 5, 10, 5).
			WillReturnRows(prevNextRows)

		// Error in tags query
		mock.ExpectQuery(`SELECT tags.tag_id, tagtype_id, tag_name FROM tagmap LEFT JOIN tags`).
			WithArgs(5).
			WillReturnError(sqlmock.ErrCancelled)

		model := ImageModel{
			Ib: 1,
			ID: 5,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for tags query failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 9: Error in tags scan
	t.Run("Error in tags scan", func(t *testing.T) {
		// Mock image info query
		imageRows := sqlmock.NewRows([]string{
			"image_id", "thread_id", "post_num", "post_id", "image_file", "image_orig_height", "image_orig_width",
		}).AddRow(5, 10, 3, 15, "test.jpg", 800, 600)

		mock.ExpectQuery(`SELECT image_id, posts.thread_id, posts.post_num, posts.post_id, image_file, image_orig_height, image_orig_width FROM images`).
			WithArgs(5, 1).
			WillReturnRows(imageRows)

		// Mock prev/next query
		prevNextRows := sqlmock.NewRows([]string{"previous", "next"}).
			AddRow(4, 6)

		mock.ExpectQuery(`SELECT \(SELECT image_id FROM images INNER JOIN posts.+\) AS previous, \(SELECT image_id.+\) AS next`).
			WithArgs(10, 5, 10, 5).
			WillReturnRows(prevNextRows)

		// Create row with wrong number of columns to cause scan error
		tagRows := sqlmock.NewRows([]string{"tag_id"}).
			AddRow(1)

		mock.ExpectQuery(`SELECT tags.tag_id, tagtype_id, tag_name FROM tagmap LEFT JOIN tags`).
			WithArgs(5).
			WillReturnRows(tagRows)

		model := ImageModel{
			Ib: 1,
			ID: 5,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for tags scan failure")
	})

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All mock expectations should be met")
}