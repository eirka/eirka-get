package models

import (
	"database/sql"
	"testing"

	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestTagModelGet(t *testing.T) {
	// Create a test DB mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")

	// Setup config for pagination
	config.Settings.Limits.PostsPerPage = 10

	// Test case 1: Valid request with images
	t.Run("Valid request with images", func(t *testing.T) {
		// Mock tag info query
		tagName := "anime"
		tagType := uint(1)
		tagRows := sqlmock.NewRows([]string{"tag_name", "tagtype_id", "count"}).
			AddRow(tagName, tagType, 25)

		mock.ExpectQuery(`SELECT tag_name, tagtype_id, COUNT\(tagmap.image_id\) FROM tags.+HAVING tag_name IS NOT NULL`).
			WithArgs(5, 1).
			WillReturnRows(tagRows)

		// Mock images query
		imageRows := sqlmock.NewRows([]string{
			"image_id", "image_file", "image_thumbnail", "image_tn_height", "image_tn_width",
		}).
			AddRow(1, "image1.jpg", "thumb1.jpg", 100, 150).
			AddRow(2, "image2.jpg", "thumb2.jpg", 120, 180).
			AddRow(3, "image3.jpg", "thumb3.jpg", 110, 160)

		mock.ExpectQuery(`SELECT images.image_id, image_file, image_thumbnail, image_tn_height, image_tn_width FROM tagmap`).
			WithArgs(5, 0, 10).
			WillReturnRows(imageRows)

		// Create model and call Get
		model := TagModel{
			Ib:   1,
			Tag:  5,
			Page: 1,
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request")

		// Validate response structure
		assert.Equal(t, uint(5), model.Result.Body.Items.(TagHeader).ID, "Tag ID should be 5")
		assert.Equal(t, &tagName, model.Result.Body.Items.(TagHeader).Tag, "Tag name should match")
		assert.Equal(t, &tagType, model.Result.Body.Items.(TagHeader).Type, "Tag type should match")

		// Check pagination
		assert.Equal(t, uint(25), model.Result.Body.Total, "Total should be 25")
		assert.Equal(t, uint(0), model.Result.Body.Limit, "Limit should be 0")
		assert.Equal(t, uint(10), model.Result.Body.PerPage, "PerPage should be 10")
		assert.Equal(t, uint(3), model.Result.Body.Pages, "Pages should be 3")
		assert.Equal(t, uint(1), model.Result.Body.CurrentPage, "CurrentPage should be 1")

		// Check images
		images := model.Result.Body.Items.(TagHeader).Images
		assert.Equal(t, 3, len(images), "Should have 3 images")

		// Check first image
		assert.Equal(t, uint(1), images[0].ID, "First image ID should be 1")
		assert.Equal(t, "image1.jpg", *images[0].File, "First image filename should match")
		assert.Equal(t, "thumb1.jpg", *images[0].Thumb, "First image thumbnail should match")
		assert.Equal(t, uint(100), *images[0].ThumbHeight, "First image thumbnail height should match")
		assert.Equal(t, uint(150), *images[0].ThumbWidth, "First image thumbnail width should match")
	})

	// Test case 2: Request with page 0 (all images)
	t.Run("Request with page 0 (all images)", func(t *testing.T) {
		// Mock tag info query
		tagName := "anime"
		tagType := uint(1)
		tagRows := sqlmock.NewRows([]string{"tag_name", "tagtype_id", "count"}).
			AddRow(tagName, tagType, 3)

		mock.ExpectQuery(`SELECT tag_name, tagtype_id, COUNT\(tagmap.image_id\) FROM tags.+HAVING tag_name IS NOT NULL`).
			WithArgs(5, 1).
			WillReturnRows(tagRows)

		// Mock images query - for page 0, limit will be 0 and perpage will be total (3)
		imageRows := sqlmock.NewRows([]string{
			"image_id", "image_file", "image_thumbnail", "image_tn_height", "image_tn_width",
		}).
			AddRow(1, "image1.jpg", "thumb1.jpg", 100, 150).
			AddRow(2, "image2.jpg", "thumb2.jpg", 120, 180).
			AddRow(3, "image3.jpg", "thumb3.jpg", 110, 160)

		mock.ExpectQuery(`SELECT images.image_id, image_file, image_thumbnail, image_tn_height, image_tn_width FROM tagmap`).
			WithArgs(5, 0, 3).
			WillReturnRows(imageRows)

		// Create model and call Get
		model := TagModel{
			Ib:   1,
			Tag:  5,
			Page: 0, // Page 0 should return all images
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for page 0 request")

		// Verify pagination settings for page 0
		assert.Equal(t, uint(3), model.Result.Body.Total, "Total images should be 3")
		assert.Equal(t, uint(3), model.Result.Body.PerPage, "PerPage should be equal to total for page 0")
		assert.Equal(t, uint(0), model.Result.Body.Limit, "Limit should be 0 for page 0")

		// Verify all images are returned
		images := model.Result.Body.Items.(TagHeader).Images
		assert.Equal(t, 3, len(images), "All 3 images should be returned for page 0")
	})

	// Test case 3: Empty parameters
	t.Run("Empty parameters", func(t *testing.T) {
		model := TagModel{
			Ib:   0,
			Tag:  0,
			Page: 1,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for empty parameters")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 4: Missing imageboard ID
	t.Run("Missing imageboard ID", func(t *testing.T) {
		model := TagModel{
			Ib:   0, // Missing imageboard ID
			Tag:  5,
			Page: 1,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for missing imageboard ID")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 5: Missing tag ID
	t.Run("Missing tag ID", func(t *testing.T) {
		model := TagModel{
			Ib:   1,
			Tag:  0, // Missing tag ID
			Page: 1,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for missing tag ID")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 6: Tag not found
	t.Run("Tag not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT tag_name, tagtype_id, COUNT\(tagmap.image_id\) FROM tags.+HAVING tag_name IS NOT NULL`).
			WithArgs(5, 1).
			WillReturnError(sql.ErrNoRows)

		model := TagModel{
			Ib:   1,
			Tag:  5,
			Page: 1,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for tag not found")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 7: Page exceeds total pages
	t.Run("Page exceeds total pages", func(t *testing.T) {
		// Mock tag info query
		tagName := "anime"
		tagType := uint(1)
		tagRows := sqlmock.NewRows([]string{"tag_name", "tagtype_id", "count"}).
			AddRow(tagName, tagType, 25)

		mock.ExpectQuery(`SELECT tag_name, tagtype_id, COUNT\(tagmap.image_id\) FROM tags.+HAVING tag_name IS NOT NULL`).
			WithArgs(5, 1).
			WillReturnRows(tagRows)

		// Page 5 would exceed total pages (3 with 25 total and 10 per page)
		model := TagModel{
			Ib:   1,
			Tag:  5,
			Page: 5,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for page exceeding total")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 8: Database connection error
	t.Run("Database connection error", func(t *testing.T) {
		// Force error by closing the mock db
		db.CloseDb()

		model := TagModel{
			Ib:   1,
			Tag:  5,
			Page: 1,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for DB connection failure")
	})

	// Test case 9: Error in tag info query
	t.Run("Error in tag info query", func(t *testing.T) {
		// Restore db connection
		mock, err = db.NewTestDb()
		assert.NoError(t, err, "An error was not expected")

		mock.ExpectQuery(`SELECT tag_name, tagtype_id, COUNT\(tagmap.image_id\) FROM tags.+HAVING tag_name IS NOT NULL`).
			WithArgs(5, 1).
			WillReturnError(sqlmock.ErrCancelled)

		model := TagModel{
			Ib:   1,
			Tag:  5,
			Page: 1,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for tag info query failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 10: Error in images query
	t.Run("Error in images query", func(t *testing.T) {
		// Mock tag info query
		tagName := "anime"
		tagType := uint(1)
		tagRows := sqlmock.NewRows([]string{"tag_name", "tagtype_id", "count"}).
			AddRow(tagName, tagType, 25)

		mock.ExpectQuery(`SELECT tag_name, tagtype_id, COUNT\(tagmap.image_id\) FROM tags.+HAVING tag_name IS NOT NULL`).
			WithArgs(5, 1).
			WillReturnRows(tagRows)

		// Error in images query
		mock.ExpectQuery(`SELECT images.image_id, image_file, image_thumbnail, image_tn_height, image_tn_width FROM tagmap`).
			WithArgs(5, 0, 10).
			WillReturnError(sqlmock.ErrCancelled)

		model := TagModel{
			Ib:   1,
			Tag:  5,
			Page: 1,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for images query failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 11: Error in image scan
	t.Run("Error in image scan", func(t *testing.T) {
		// Mock tag info query
		tagName := "anime"
		tagType := uint(1)
		tagRows := sqlmock.NewRows([]string{"tag_name", "tagtype_id", "count"}).
			AddRow(tagName, tagType, 25)

		mock.ExpectQuery(`SELECT tag_name, tagtype_id, COUNT\(tagmap.image_id\) FROM tags.+HAVING tag_name IS NOT NULL`).
			WithArgs(5, 1).
			WillReturnRows(tagRows)

		// Create row with wrong number of columns to cause scan error
		imageRows := sqlmock.NewRows([]string{
			"image_id", "image_file", // Missing columns
		}).AddRow(1, "image1.jpg")

		mock.ExpectQuery(`SELECT images.image_id, image_file, image_thumbnail, image_tn_height, image_tn_width FROM tagmap`).
			WithArgs(5, 0, 10).
			WillReturnRows(imageRows)

		model := TagModel{
			Ib:   1,
			Tag:  5,
			Page: 1,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for image scan failure")
	})

	// Test case 12: Valid request with no images
	t.Run("Valid request with no images", func(t *testing.T) {
		// Mock tag info query
		tagName := "rare_tag"
		tagType := uint(2)
		tagRows := sqlmock.NewRows([]string{"tag_name", "tagtype_id", "count"}).
			AddRow(tagName, tagType, 0)

		mock.ExpectQuery(`SELECT tag_name, tagtype_id, COUNT\(tagmap.image_id\) FROM tags.+HAVING tag_name IS NOT NULL`).
			WithArgs(5, 1).
			WillReturnRows(tagRows)

		// Mock empty images query
		imageRows := sqlmock.NewRows([]string{
			"image_id", "image_file", "image_thumbnail", "image_tn_height", "image_tn_width",
		})

		mock.ExpectQuery(`SELECT images.image_id, image_file, image_thumbnail, image_tn_height, image_tn_width FROM tagmap`).
			WithArgs(5, 0, 10).
			WillReturnRows(imageRows)

		model := TagModel{
			Ib:   1,
			Tag:  5,
			Page: 1,
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request with no images")

		// Validate response structure
		assert.Equal(t, uint(5), model.Result.Body.Items.(TagHeader).ID, "Tag ID should be 5")
		assert.Equal(t, &tagName, model.Result.Body.Items.(TagHeader).Tag, "Tag name should match")
		assert.Equal(t, &tagType, model.Result.Body.Items.(TagHeader).Type, "Tag type should match")
		assert.Equal(t, uint(0), model.Result.Body.Total, "Total should be 0")
		assert.Equal(t, 0, len(model.Result.Body.Items.(TagHeader).Images), "Should have 0 images")
	})

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All mock expectations should be met")
}