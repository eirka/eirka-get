package models

import (
	"testing"

	"github.com/eirka/eirka-libs/db"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestTagTypesModelGet(t *testing.T) {
	// Create a test DB mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")

	// Test case 1: Valid request with multiple tag types
	t.Run("Valid request with multiple tag types", func(t *testing.T) {
		// Mock tag types query
		tagTypeRows := sqlmock.NewRows([]string{"tagtype_id", "tagtype_name"}).
			AddRow(1, "character").
			AddRow(2, "copyright").
			AddRow(3, "artist").
			AddRow(4, "genre")

		mock.ExpectQuery(`select tagtype_id,tagtype_name from tagtype`).
			WillReturnRows(tagTypeRows)

		// Create model and call Get
		model := TagTypesModel{}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request")

		// Validate response structure
		tagTypes := model.Result.Body
		assert.Equal(t, 4, len(tagTypes), "Should have 4 tag types")

		// Check first tag type
		assert.Equal(t, uint(1), tagTypes[0].ID, "First tag type ID should be 1")
		assert.Equal(t, "character", tagTypes[0].Type, "First tag type should match")

		// Check second tag type
		assert.Equal(t, uint(2), tagTypes[1].ID, "Second tag type ID should be 2")
		assert.Equal(t, "copyright", tagTypes[1].Type, "Second tag type should match")

		// Check third tag type
		assert.Equal(t, uint(3), tagTypes[2].ID, "Third tag type ID should be 3")
		assert.Equal(t, "artist", tagTypes[2].Type, "Third tag type should match")

		// Check fourth tag type
		assert.Equal(t, uint(4), tagTypes[3].ID, "Fourth tag type ID should be 4")
		assert.Equal(t, "genre", tagTypes[3].Type, "Fourth tag type should match")
	})

	// Test case 2: Valid request with empty result
	t.Run("Valid request with empty result", func(t *testing.T) {
		// Mock empty tag types query
		tagTypeRows := sqlmock.NewRows([]string{"tagtype_id", "tagtype_name"})

		mock.ExpectQuery(`select tagtype_id,tagtype_name from tagtype`).
			WillReturnRows(tagTypeRows)

		// Create model and call Get
		model := TagTypesModel{}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request with empty result")

		// Validate response structure
		tagTypes := model.Result.Body
		assert.Equal(t, 0, len(tagTypes), "Should have 0 tag types")
	})

	// Test case 3: Database connection error
	t.Run("Database connection error", func(t *testing.T) {
		// Force error by closing the mock db
		db.CloseDb()

		model := TagTypesModel{}

		err := model.Get()
		assert.Error(t, err, "Should return error for DB connection failure")
	})

	// Test case 4: Query execution error
	t.Run("Query execution error", func(t *testing.T) {
		// Restore db connection
		mock, err = db.NewTestDb()
		assert.NoError(t, err, "An error was not expected")

		mock.ExpectQuery(`select tagtype_id,tagtype_name from tagtype`).
			WillReturnError(sqlmock.ErrCancelled)

		model := TagTypesModel{}

		err = model.Get()
		assert.Error(t, err, "Should return error for query execution failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 5: Error in scan
	t.Run("Error in scan", func(t *testing.T) {
		// Create row with wrong number of columns to cause scan error
		tagTypeRows := sqlmock.NewRows([]string{"tagtype_id"}).
			AddRow(1)

		mock.ExpectQuery(`select tagtype_id,tagtype_name from tagtype`).
			WillReturnRows(tagTypeRows)

		model := TagTypesModel{}

		err = model.Get()
		assert.Error(t, err, "Should return error for scan failure")
	})

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All mock expectations should be met")
}