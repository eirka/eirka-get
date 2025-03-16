package models_test

import (
	"testing"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/eirka/eirka-get/models"
)

func TestFavoritedModelGet(t *testing.T) {
	// Create a test DB mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")

	// Test case 1: Valid request with favorited images
	t.Run("Valid request with favorited images", func(t *testing.T) {
		// Mock favorited images query
		rows := sqlmock.NewRows([]string{
			"image_id", "image_file", "image_thumbnail", "image_tn_height", "image_tn_width",
		}).
			AddRow(1, "image1.jpg", "thumb1.jpg", 150, 100).
			AddRow(2, "image2.jpg", "thumb2.jpg", 200, 150)

		mock.ExpectQuery(`SELECT image_id, image_file, image_thumbnail, image_tn_height, image_tn_width FROM.*`).
			WithArgs(1).
			WillReturnRows(rows)

		// Create model and call Get
		model := models.FavoritedModel{
			Ib: 1,
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request")

		// Validate response structure
		assert.Equal(t, 2, len(model.Result.Body), "Should have 2 favorited images")
		
		// Check first image data
		assert.Equal(t, uint(1), model.Result.Body[0].ID, "First image ID should be 1")
		assert.Equal(t, "image1.jpg", *model.Result.Body[0].File, "First image filename should match")
		assert.Equal(t, "thumb1.jpg", *model.Result.Body[0].Thumb, "First thumbnail should match")
		assert.Equal(t, uint(150), *model.Result.Body[0].ThumbHeight, "First thumbnail height should match")
		assert.Equal(t, uint(100), *model.Result.Body[0].ThumbWidth, "First thumbnail width should match")
		
		// Check second image data
		assert.Equal(t, uint(2), model.Result.Body[1].ID, "Second image ID should be 2")
		assert.Equal(t, "image2.jpg", *model.Result.Body[1].File, "Second image filename should match")
		assert.Equal(t, "thumb2.jpg", *model.Result.Body[1].Thumb, "Second thumbnail should match")
		assert.Equal(t, uint(200), *model.Result.Body[1].ThumbHeight, "Second thumbnail height should match")
		assert.Equal(t, uint(150), *model.Result.Body[1].ThumbWidth, "Second thumbnail width should match")
	})

	// Test case 2: Valid request but no favorited images found
	t.Run("Valid request with no favorited images", func(t *testing.T) {
		// Mock favorited images query with empty result
		rows := sqlmock.NewRows([]string{
			"image_id", "image_file", "image_thumbnail", "image_tn_height", "image_tn_width",
		})

		mock.ExpectQuery(`SELECT image_id, image_file, image_thumbnail, image_tn_height, image_tn_width FROM.*`).
			WithArgs(1).
			WillReturnRows(rows)

		// Create model and call Get
		model := models.FavoritedModel{
			Ib: 1,
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request with no results")
		assert.Empty(t, model.Result.Body, "Image list should be empty")
	})

	// Test case 3: Empty parameter (image board ID is 0)
	t.Run("Empty parameter", func(t *testing.T) {
		model := models.FavoritedModel{
			Ib: 0,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for empty parameter")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 4: Database connection error
	t.Run("Database connection error", func(t *testing.T) {
		// Force error by closing the mock db
		db.CloseDb()

		model := models.FavoritedModel{
			Ib: 1,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for DB connection failure")
	})

	// Test case 5: Error in query
	t.Run("Error in query", func(t *testing.T) {
		// Restore db connection
		mock, err = db.NewTestDb()
		assert.NoError(t, err, "An error was not expected")

		mock.ExpectQuery(`SELECT image_id, image_file, image_thumbnail, image_tn_height, image_tn_width FROM.*`).
			WithArgs(1).
			WillReturnError(sqlmock.ErrCancelled)

		model := models.FavoritedModel{
			Ib: 1,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for query failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 6: Error in scan
	t.Run("Error in scan", func(t *testing.T) {
		// Create row with wrong number of columns to cause scan error
		rows := sqlmock.NewRows([]string{
			"image_id", "image_file", // Missing columns
		}).AddRow(1, "image1.jpg")

		mock.ExpectQuery(`SELECT image_id, image_file, image_thumbnail, image_tn_height, image_tn_width FROM.*`).
			WithArgs(1).
			WillReturnRows(rows)

		model := models.FavoritedModel{
			Ib: 1,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for scan failure")
	})

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All mock expectations should be met")
}