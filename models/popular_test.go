package models_test

import (
	"testing"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/eirka/eirka-get/models"
)

func TestPopularModelGet(t *testing.T) {
	// Create a test DB mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")

	// Test case 1: Valid request with popular images
	t.Run("Valid request with popular images", func(t *testing.T) {
		// Mock popular images query
		rows := sqlmock.NewRows([]string{
			"request_itemvalue", "image_file", "image_thumbnail", "image_tn_height", "image_tn_width",
		}).
			AddRow(1, "image1.jpg", "thumb1.jpg", 150, 100).
			AddRow(2, "image2.jpg", "thumb2.jpg", 200, 150).
			AddRow(3, "image3.jpg", "thumb3.jpg", 250, 180)

		mock.ExpectQuery(`SELECT request_itemvalue, image_file, image_thumbnail, image_tn_height, image_tn_width FROM.*`).
			WithArgs(1).
			WillReturnRows(rows)

		// Create model and call Get
		model := models.PopularModel{
			Ib: 1,
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request")

		// Validate response structure
		assert.Equal(t, 3, len(model.Result.Body), "Should have 3 popular images")
		
		// Check first image data
		assert.Equal(t, uint(1), model.Result.Body[0].ID, "First image ID should be 1")
		assert.Equal(t, "image1.jpg", *model.Result.Body[0].File, "First image filename should match")
		assert.Equal(t, "thumb1.jpg", *model.Result.Body[0].Thumb, "First thumbnail should match")
		assert.Equal(t, uint(150), *model.Result.Body[0].ThumbHeight, "First thumbnail height should match")
		assert.Equal(t, uint(100), *model.Result.Body[0].ThumbWidth, "First thumbnail width should match")
		
		// Check last image data
		assert.Equal(t, uint(3), model.Result.Body[2].ID, "Last image ID should be 3")
		assert.Equal(t, "image3.jpg", *model.Result.Body[2].File, "Last image filename should match")
		assert.Equal(t, "thumb3.jpg", *model.Result.Body[2].Thumb, "Last thumbnail should match")
		assert.Equal(t, uint(250), *model.Result.Body[2].ThumbHeight, "Last thumbnail height should match")
		assert.Equal(t, uint(180), *model.Result.Body[2].ThumbWidth, "Last thumbnail width should match")
	})

	// Test case 2: Valid request but no popular images found
	t.Run("Valid request with no popular images", func(t *testing.T) {
		// Mock popular images query with empty result
		rows := sqlmock.NewRows([]string{
			"request_itemvalue", "image_file", "image_thumbnail", "image_tn_height", "image_tn_width",
		})

		mock.ExpectQuery(`SELECT request_itemvalue, image_file, image_thumbnail, image_tn_height, image_tn_width FROM.*`).
			WithArgs(1).
			WillReturnRows(rows)

		// Create model and call Get
		model := models.PopularModel{
			Ib: 1,
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request with no results")
		assert.Empty(t, model.Result.Body, "Image list should be empty")
	})

	// Test case 3: Empty parameter (image board ID is 0)
	t.Run("Empty parameter", func(t *testing.T) {
		model := models.PopularModel{
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

		model := models.PopularModel{
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

		mock.ExpectQuery(`SELECT request_itemvalue, image_file, image_thumbnail, image_tn_height, image_tn_width FROM.*`).
			WithArgs(1).
			WillReturnError(sqlmock.ErrCancelled)

		model := models.PopularModel{
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
			"request_itemvalue", "image_file", // Missing columns
		}).AddRow(1, "image1.jpg")

		mock.ExpectQuery(`SELECT request_itemvalue, image_file, image_thumbnail, image_tn_height, image_tn_width FROM.*`).
			WithArgs(1).
			WillReturnRows(rows)

		model := models.PopularModel{
			Ib: 1,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for scan failure")
	})

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All mock expectations should be met")
}