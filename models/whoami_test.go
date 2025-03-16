package models_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/eirka/eirka-libs/user"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/eirka/eirka-get/models"
)

func TestWhoAmIModelGet(t *testing.T) {
	// Create a test DB mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")

	// Create a couple of test user objects
	authenticatedUser := user.User{
		ID:              1,
		IsAuthenticated: true,
	}

	unauthenticatedUser := user.User{
		ID:              2,
		IsAuthenticated: false,
	}

	// Test case 1: Valid request with authenticated user and activity history
	t.Run("Valid request with authenticated user and activity history", func(t *testing.T) {
		email := "user@example.com"
		lastActiveTime := time.Now().Add(-24 * time.Hour) // 1 day ago

		// Mock user info query
		userRows := sqlmock.NewRows([]string{"role", "user_name", "user_email"}).
			AddRow(3, "testuser", email)

		mock.ExpectQuery(`SELECT COALESCE.*`).
			WithArgs(1, 1).
			WillReturnRows(userRows)

		// Mock last active query
		activeRows := sqlmock.NewRows([]string{"request_time"}).
			AddRow(lastActiveTime)

		mock.ExpectQuery(`SELECT request_time FROM analytics WHERE user_id = \? AND ib_id = \? ORDER BY analytics_id DESC LIMIT 1`).
			WithArgs(1, 1).
			WillReturnRows(activeRows)

		// Create model and call Get
		model := models.WhoAmIModel{
			User: authenticatedUser,
			Ib:   1,
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request")

		// Validate response structure
		assert.Equal(t, uint(1), model.Result.Body.ID, "User ID should be 1")
		assert.Equal(t, "testuser", model.Result.Body.Name, "Username should match")
		assert.Equal(t, uint(3), model.Result.Body.Group, "User group should be 3")
		assert.Equal(t, true, model.Result.Body.Authenticated, "User should be authenticated")
		assert.NotNil(t, model.Result.Body.Email, "Email should not be nil")
		assert.Equal(t, email, *model.Result.Body.Email, "Email should match")
		assert.Equal(t, lastActiveTime.Unix(), model.Result.Body.LastActive.Unix(), "Last active time should match")
	})

	// Test case 2: Valid request with authenticated user but no activity history
	t.Run("Valid request with authenticated user but no activity history", func(t *testing.T) {
		email := "user@example.com"

		// Mock user info query
		userRows := sqlmock.NewRows([]string{"role", "user_name", "user_email"}).
			AddRow(3, "testuser", email)

		mock.ExpectQuery(`SELECT COALESCE.*`).
			WithArgs(1, 1).
			WillReturnRows(userRows)

		// Mock last active query with no rows
		mock.ExpectQuery(`SELECT request_time FROM analytics WHERE user_id = \? AND ib_id = \? ORDER BY analytics_id DESC LIMIT 1`).
			WithArgs(1, 1).
			WillReturnError(sql.ErrNoRows)

		// Create model and call Get
		model := models.WhoAmIModel{
			User: authenticatedUser,
			Ib:   1,
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request with no activity history")

		// Validate response
		assert.Equal(t, uint(1), model.Result.Body.ID, "User ID should be 1")
		assert.True(t, model.Result.Body.Authenticated, "User should be authenticated")
		// LastActive should be approximately current time since no history found
		assert.WithinDuration(t, time.Now(), model.Result.Body.LastActive, 2*time.Second, "Last active should be current time")
	})

	// Test case 3: Valid request with unauthenticated user
	t.Run("Valid request with unauthenticated user", func(t *testing.T) {
		email := "user@example.com"

		// Mock user info query
		userRows := sqlmock.NewRows([]string{"role", "user_name", "user_email"}).
			AddRow(1, "guest", email)

		mock.ExpectQuery(`SELECT COALESCE.*`).
			WithArgs(1, 2).
			WillReturnRows(userRows)

		// Create model and call Get
		model := models.WhoAmIModel{
			User: unauthenticatedUser,
			Ib:   1,
		}

		err := model.Get()
		assert.NoError(t, err, "No error should be returned for valid request with unauthenticated user")

		// Validate response
		assert.Equal(t, uint(2), model.Result.Body.ID, "User ID should be 2")
		assert.Equal(t, "guest", model.Result.Body.Name, "Username should be guest")
		assert.Equal(t, uint(1), model.Result.Body.Group, "User group should be 1")
		assert.False(t, model.Result.Body.Authenticated, "User should not be authenticated")
		assert.NotNil(t, model.Result.Body.Email, "Email should not be nil")
		// LastActive should be approximately current time for unauthenticated user
		assert.WithinDuration(t, time.Now(), model.Result.Body.LastActive, 2*time.Second, "Last active should be current time")
	})

	// Test case 4: Empty parameters
	t.Run("Empty parameters", func(t *testing.T) {
		emptyUser := user.User{
			ID:              0,
			IsAuthenticated: false,
		}

		model := models.WhoAmIModel{
			User: emptyUser,
			Ib:   0,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for empty parameters")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 5: Missing User ID
	t.Run("Missing User ID", func(t *testing.T) {
		emptyUser := user.User{
			ID:              0,
			IsAuthenticated: true,
		}

		model := models.WhoAmIModel{
			User: emptyUser,
			Ib:   1,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for missing User ID")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 6: Missing Imageboard ID
	t.Run("Missing Imageboard ID", func(t *testing.T) {
		model := models.WhoAmIModel{
			User: authenticatedUser,
			Ib:   0,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for missing Imageboard ID")
		assert.Equal(t, e.ErrNotFound, err, "Should return ErrNotFound")
	})

	// Test case 7: Database connection error
	t.Run("Database connection error", func(t *testing.T) {
		// Force error by closing the mock db
		db.CloseDb()

		model := models.WhoAmIModel{
			User: authenticatedUser,
			Ib:   1,
		}

		err := model.Get()
		assert.Error(t, err, "Should return error for DB connection failure")
	})

	// Test case 8: Error in user info query
	t.Run("Error in user info query", func(t *testing.T) {
		// Restore db connection
		mock, err = db.NewTestDb()
		assert.NoError(t, err, "An error was not expected")

		mock.ExpectQuery(`SELECT COALESCE.*`).
			WithArgs(1, 1).
			WillReturnError(sqlmock.ErrCancelled)

		model := models.WhoAmIModel{
			User: authenticatedUser,
			Ib:   1,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for user info query failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 9: Error in analytics query
	t.Run("Error in analytics query", func(t *testing.T) {
		email := "user@example.com"

		// Mock user info query
		userRows := sqlmock.NewRows([]string{"role", "user_name", "user_email"}).
			AddRow(3, "testuser", email)

		mock.ExpectQuery(`SELECT COALESCE.*`).
			WithArgs(1, 1).
			WillReturnRows(userRows)

		// Error in analytics query
		mock.ExpectQuery(`SELECT request_time FROM analytics WHERE user_id = \? AND ib_id = \? ORDER BY analytics_id DESC LIMIT 1`).
			WithArgs(1, 1).
			WillReturnError(sqlmock.ErrCancelled)

		model := models.WhoAmIModel{
			User: authenticatedUser,
			Ib:   1,
		}

		err = model.Get()
		assert.Error(t, err, "Should return error for analytics query failure")
		assert.Equal(t, sqlmock.ErrCancelled, err, "Should return the SQL error")
	})

	// Test case 10: Last active time is null
	t.Run("Last active time is null", func(t *testing.T) {
		email := "user@example.com"

		// Mock user info query
		userRows := sqlmock.NewRows([]string{"role", "user_name", "user_email"}).
			AddRow(3, "testuser", email)

		mock.ExpectQuery(`SELECT COALESCE.*`).
			WithArgs(1, 1).
			WillReturnRows(userRows)

		// Mock last active query with null value
		activeRows := sqlmock.NewRows([]string{"request_time"}).
			AddRow(nil)

		mock.ExpectQuery(`SELECT request_time FROM analytics WHERE user_id = \? AND ib_id = \? ORDER BY analytics_id DESC LIMIT 1`).
			WithArgs(1, 1).
			WillReturnRows(activeRows)

		model := models.WhoAmIModel{
			User: authenticatedUser,
			Ib:   1,
		}

		err = model.Get()
		assert.NoError(t, err, "No error should be returned when last active time is null")

		// LastActive should be approximately current time
		assert.WithinDuration(t, time.Now(), model.Result.Body.LastActive, 2*time.Second, "Last active should be current time when null")
	})

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All mock expectations should be met")
}
