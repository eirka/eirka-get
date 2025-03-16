package models_test

import (
	"testing"

	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/eirka/eirka-get/models"
)

func TestFavoritesModelGet(t *testing.T) {
	// Create a test DB mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")

	// Setup config for pagination
	config.Settings.Limits.PostsPerPage = 10

	// Test case 1: Valid request with multiple favorited images
	t.Run("Valid request with multiple favorited images", func(t *testing.T) {
		// Mock favorites count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(15)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorites INNER JOIN images.*`).
			WithArgs(1, 1).
			WillReturnRows(countRows)

		// Mock favorited images query
		rows := sqlmock.NewRows([]string{
			"image_id", "image_file", "image_thumbnail", "image_tn_height", "image_tn_width",
		}).
			AddRow(5, "image5.jpg", "thumb5.jpg", 150, 100).
			AddRow(4, "image4.jpg", "thumb4.jpg", 200, 150).
			AddRow(3, "image3.jpg", "thumb3.jpg", 250, 180)

		mock.ExpectQuery(`SELECT images.image_id, image_file, image_thumbnail, image_tn_height, image_tn_width FROM favorites INNER JOIN.*`).
			WithArgs(1, 1, 0, 10).
			WillReturnRows(rows)

		// Create model and call Get
		model := models.FavoritesModel{
			User: 1,
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
		header, ok := model.Result.Body.Items.(models.FavoritesHeader)
		assert.True(t, ok, "Items should be of type models.FavoritesHeader")
		assert.Equal(t, 3, len(header.Images), "Should have 3 favorite images")
		
		// Check first image data
		assert.Equal(t, uint(5), header.Images[0].ID, "First image ID should be 5")
		assert.Equal(t, "image5.jpg", *header.Images[0].File, "First image filename should match")
		assert.Equal(t, "thumb5.jpg", *header.Images[0].Thumb, "First thumbnail should match")
		assert.Equal(t, uint(150), *header.Images[0].ThumbHeight, "First thumbnail height should match")
		assert.Equal(t, uint(100), *header.Images[0].ThumbWidth, "First thumbnail width should match")
	})

	// Test case 2: Valid request with no favorited images
	t.Run("Valid request with no favorited images", func(t *testing.T) {
		// Mock favorites count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorites INNER JOIN images.*`).
			WithArgs(1, 1).
			WillReturnRows(countRows)

		// Mock favorited images query with empty result
		rows := sqlmock.NewRows([]string{
			"image_id", "image_file", "image_thumbnail", "image_tn_height", "image_tn_width",
		})

		mock.ExpectQuery(`SELECT images.image_id, image_file, image_thumbnail, image_tn_height, image_tn_width FROM favorites INNER JOIN.*`).
			WithArgs(1, 1, 0, 10).
			WillReturnRows(rows)

		// Create model and call Get
		model := models.FavoritesModel{
			User: 1,
			Ib:   1,
			Page: 1,
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request with no results")
		
		// Validate response structure
		assert.Equal(t, uint(0), model.Result.Body.Total, "Total should be 0")
		assert.Equal(t, uint(1), model.Result.Body.Pages, "Pages should be 1")
		
		// Check items type and content
		header, ok := model.Result.Body.Items.(models.FavoritesHeader)
		assert.True(t, ok, "Items should be of type models.FavoritesHeader")
		assert.Empty(t, header.Images, "Image list should be empty")
	})

	// Test case 3: Empty parameters
	t.Run("Empty parameters", func(t *testing.T) {
		model := models.FavoritesModel{
			User: 0,
			Ib:   0,
			Page: 0,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for empty parameters")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 4: Missing User ID
	t.Run("Missing User ID", func(t *testing.T) {
		model := models.FavoritesModel{
			User: 0,
			Ib:   1,
			Page: 1,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for missing User ID")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 5: Missing Imageboard ID
	t.Run("Missing Imageboard ID", func(t *testing.T) {
		model := models.FavoritesModel{
			User: 1,
			Ib:   0,
			Page: 1,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for missing Imageboard ID")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 6: Missing Page
	t.Run("Missing Page", func(t *testing.T) {
		model := models.FavoritesModel{
			User: 1,
			Ib:   1,
			Page: 0,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for missing Page")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 7: Page exceeds total pages
	t.Run("Page exceeds total pages", func(t *testing.T) {
		// Mock favorites count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(15)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorites INNER JOIN images.*`).
			WithArgs(1, 1).
			WillReturnRows(countRows)

		model := models.FavoritesModel{
			User: 1,
			Ib:   1,
			Page: 5, // Should exceed total pages which is 2
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for page exceeding total")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 8: Database connection error
	t.Run("Database connection error", func(t *testing.T) {
		// Force error by closing the mock db
		db.CloseDb()

		model := models.FavoritesModel{
			User: 1,
			Ib:   1,
			Page: 1,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for DB connection failure")
	})

	// Test case 9: Error in count query
	t.Run("Error in count query", func(t *testing.T) {
		// Restore db connection
		mock, err = db.NewTestDb()
		assert.NoError(t, err, "An error was not expected")

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorites INNER JOIN images.*`).
			WithArgs(1, 1).
			WillReturnError(sqlmock.ErrCancelled)

		model := models.FavoritesModel{
			User: 1,
			Ib:   1,
			Page: 1,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for count query failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 10: Error in images query
	t.Run("Error in images query", func(t *testing.T) {
		// Mock favorites count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(15)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorites INNER JOIN images.*`).
			WithArgs(1, 1).
			WillReturnRows(countRows)

		mock.ExpectQuery(`SELECT images.image_id, image_file, image_thumbnail, image_tn_height, image_tn_width FROM favorites INNER JOIN.*`).
			WithArgs(1, 1, 0, 10).
			WillReturnError(sqlmock.ErrCancelled)

		model := models.FavoritesModel{
			User: 1,
			Ib:   1,
			Page: 1,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for images query failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 11: Error in scan
	t.Run("Error in scan", func(t *testing.T) {
		// Mock favorites count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(15)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorites INNER JOIN images.*`).
			WithArgs(1, 1).
			WillReturnRows(countRows)

		// Create row with wrong number of columns to cause scan error
		rows := sqlmock.NewRows([]string{
			"image_id", "image_file", // Missing columns
		}).AddRow(1, "image1.jpg")

		mock.ExpectQuery(`SELECT images.image_id, image_file, image_thumbnail, image_tn_height, image_tn_width FROM favorites INNER JOIN.*`).
			WithArgs(1, 1, 0, 10).
			WillReturnRows(rows)

		model := models.FavoritesModel{
			User: 1,
			Ib:   1,
			Page: 1,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for scan failure")
	})

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All mock expectations should be met")
}