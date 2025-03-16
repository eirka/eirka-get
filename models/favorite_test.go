package models

import (
	"database/sql"
	"testing"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestFavoriteModelGet(t *testing.T) {
	// Create a test DB mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")

	// Test case 1: Valid request with favorited image
	t.Run("Valid request with favorited image", func(t *testing.T) {
		// Mock the EXISTS query for favorited image (returns 1 for true)
		favRows := sqlmock.NewRows([]string{"exists"}).AddRow(1)
		mock.ExpectQuery(`SELECT EXISTS\(\s*SELECT 1\s*FROM favorites\s*WHERE user_id = \?\s*AND image_id = \?\s*\)`).
			WithArgs(1, 5).
			WillReturnRows(favRows)

		// Create model and call Get
		model := FavoriteModel{
			User: 1,
			ID:   5,
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request")

		// Validate response structure
		assert.True(t, model.Result.Starred, "Image should be marked as starred")
	})

	// Test case 2: Valid request with non-favorited image
	t.Run("Valid request with non-favorited image", func(t *testing.T) {
		// Mock the EXISTS query for non-favorited image (returns 0 for false)
		favRows := sqlmock.NewRows([]string{"exists"}).AddRow(0)
		mock.ExpectQuery(`SELECT EXISTS\(\s*SELECT 1\s*FROM favorites\s*WHERE user_id = \?\s*AND image_id = \?\s*\)`).
			WithArgs(1, 6).
			WillReturnRows(favRows)

		// Create model and call Get
		model := FavoriteModel{
			User: 1,
			ID:   6,
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request")

		// Validate response structure
		assert.False(t, model.Result.Starred, "Image should not be marked as starred")
	})

	// Test case 3: Empty parameters
	t.Run("Empty parameters", func(t *testing.T) {
		model := FavoriteModel{
			User: 0,
			ID:   0,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for empty parameters")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 4: Only User ID missing
	t.Run("Missing User ID", func(t *testing.T) {
		model := FavoriteModel{
			User: 0, // Missing user ID
			ID:   5,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for missing User ID")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 5: Only Image ID missing
	t.Run("Missing Image ID", func(t *testing.T) {
		model := FavoriteModel{
			User: 1,
			ID:   0, // Missing image ID
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for missing Image ID")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 6: Database connection error
	t.Run("Database connection error", func(t *testing.T) {
		// Force error by closing the mock db
		db.CloseDb()

		model := FavoriteModel{
			User: 1,
			ID:   5,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for DB connection failure")
	})

	// Test case 7: Error in query execution
	t.Run("Error in query execution", func(t *testing.T) {
		// Restore db connection
		mock, err = db.NewTestDb()
		assert.NoError(t, err, "An error was not expected")

		mock.ExpectQuery(`SELECT EXISTS\(\s*SELECT 1\s*FROM favorites\s*WHERE user_id = \?\s*AND image_id = \?\s*\)`).
			WithArgs(1, 5).
			WillReturnError(sqlmock.ErrCancelled)

		model := FavoriteModel{
			User: 1,
			ID:   5,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for query execution failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 8: SQL.ErrNoRows (which should be handled specially)
	t.Run("SQL.ErrNoRows", func(t *testing.T) {
		mock.ExpectQuery(`SELECT EXISTS\(\s*SELECT 1\s*FROM favorites\s*WHERE user_id = \?\s*AND image_id = \?\s*\)`).
			WithArgs(1, 5).
			WillReturnError(sql.ErrNoRows)

		model := FavoriteModel{
			User: 1,
			ID:   5,
		}

		err = model.Get()
		assert.NoError(t, err, "No error should be returned for sql.ErrNoRows")
		assert.False(t, model.Result.Starred, "Image should not be marked as starred")
	})

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All mock expectations should be met")
}