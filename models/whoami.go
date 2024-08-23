package models

import (
	"database/sql"
	"time"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/eirka/eirka-libs/user"
)

// WhoAmIModel holds the parameters from the request and also the key for the cache
type WhoAmIModel struct {
	User   user.User
	Ib     uint
	Result UserType
}

// UserType is the top level of the JSON response
type UserType struct {
	Body UserInfo `json:"user"`
}

// UserInfo holds all the user metadata
type UserInfo struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	Group         uint      `json:"group"`
	Authenticated bool      `json:"authenticated"`
	Email         *string   `json:"email,omitempty"`
	LastActive    time.Time `json:"last_active,omitempty"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *WhoAmIModel) Get() (err error) {

	if i.Ib == 0 || i.User.ID == 0 {
		return e.ErrNotFound
	}

	// Initialize response header
	response := UserType{}

	r := UserInfo{
		ID:            i.User.ID,
		Authenticated: i.User.IsAuthenticated,
	}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// This query retrieves the user's role, name, and email for the given image board (ib_id).
	// It uses COALESCE to get the highest role_id from user_ib_role_map for the specific image board,
	// falling back to the global role_id from user_role_map if no board-specific role is found.
	err = dbase.QueryRow(`
		SELECT
			COALESCE(
				(SELECT MAX(role_id)
				 FROM user_ib_role_map
				 WHERE user_ib_role_map.user_id = users.user_id
				 AND ib_id = ?),
				user_role_map.role_id
			) AS role,
			user_name,
			user_email
		FROM users
		INNER JOIN user_role_map ON (user_role_map.user_id = users.user_id)
		WHERE users.user_id = ?
	`, i.Ib, r.ID).Scan(&r.Group, &r.Name, &r.Email)
	if err != nil {
		return
	}

	// get the last time the user was active if authed
	if !r.Authenticated {
		r.LastActive = time.Now()
	} else {
		var lastactive sql.NullTime

		// This query retrieves the most recent request_time from the analytics table
		// for the given user_id and image board (ib_id), which represents the user's last active time.
		err = dbase.QueryRow(`
			SELECT request_time
			FROM analytics
			WHERE user_id = ?
			AND ib_id = ?
			ORDER BY analytics_id DESC
			LIMIT 1
		`, r.ID, i.Ib).Scan(&lastactive)
		// we dont care if there were no rows
		if err == sql.ErrNoRows {
			err = nil
		} else if err != nil {
			return
		}

		// set the last active time
		if lastactive.Valid {
			r.LastActive = lastactive.Time
		} else {
			r.LastActive = time.Now()
		}
	}

	response.Body = r

	// This is the data we will serialize
	i.Result = response

	return

}
