package models

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"time"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

// UserModel holds the parameters from the request and also the key for the cache
type UserModel struct {
	User   uint
	Ib     uint
	Result UserType
}

// UserType is the top level of the JSON response
type UserType struct {
	Body UserInfo `json:"user"`
}

// UserInfo holds all the user metadata
type UserInfo struct {
	Id         uint      `json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Group      uint      `json:"group"`
	LastActive time.Time `json:"last_active"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *UserModel) Get() (err error) {

	if i.Ib == 0 || i.User == 0 || i.User == 1 {
		return e.ErrNotFound
	}

	// Initialize response header
	response := UserType{}

	r := UserInfo{}

	// set our user id
	r.Id = i.User

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// get data from users table
	err = dbase.QueryRow(`SELECT COALESCE((SELECT MAX(role_id) FROM user_ib_role_map WHERE user_ib_role_map.user_id = users.user_id AND ib_id = ?),user_role_map.role_id) as role,
    user_name,user_email FROM users
    INNER JOIN user_role_map ON (user_role_map.user_id = users.user_id)
    WHERE users.user_id = ?`, i.Ib, i.User).Scan(&r.Group, &r.Name, &r.Email)
	if err != nil {
		return
	}

	var lastactive mysql.NullTime

	// get the time the user was last active
	err = dbase.QueryRow(`SELECT request_time FROM analytics 
    WHERE user_id = ? AND ib_id = ? ORDER BY analytics_id DESC LIMIT 1`, i.User, i.Ib).Scan(&lastactive)
	if err != nil && err != sql.ErrNoRows {
		return
	}

	// set the last active time
	if lastactive.Valid {
		r.LastActive = lastactive.Time
	} else {
		r.LastActive = time.Now()
	}

	response.Body = r

	// This is the data we will serialize
	i.Result = response

	return

}
