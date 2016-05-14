package models

import (
	"database/sql"
	"time"

	"github.com/go-sql-driver/mysql"

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
	ID         uint      `json:"id"`
	Name       string    `json:"name"`
	Group      uint      `json:"group"`
	Email      *string   `json:"email,omitempty"`
	LastActive time.Time `json:"last_active,omitempty"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *WhoAmIModel) Get() (err error) {

	if i.Ib == 0 || i.User.ID == 0 {
		return e.ErrNotFound
	}

	// Initialize response header
	response := UserType{}

	r := UserInfo{}

	// set our user id
	r.ID = i.User.ID

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// get data from users table
	err = dbase.QueryRow(`SELECT
  COALESCE((SELECT MAX(role_id) FROM user_ib_role_map WHERE user_ib_role_map.user_id = users.user_id AND ib_id = ?),user_role_map.role_id) as role,
  user_name,user_email
  FROM users
  INNER JOIN user_role_map ON (user_role_map.user_id = users.user_id)
  WHERE users.user_id = ?`, i.Ib, i.User.ID).Scan(&r.Group, &r.Name, &r.Email)
	if err != nil {
		return
	}

	// get the last time the user was active if authed
	if i.User.IsAuthenticated {

		var lastactive mysql.NullTime

		// get the time the user was last active
		err = dbase.QueryRow(`SELECT request_time FROM analytics
    WHERE user_id = ? AND ib_id = ? ORDER BY analytics_id DESC LIMIT 1`, i.User, i.Ib).Scan(&lastactive)
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
