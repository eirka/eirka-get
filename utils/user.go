package utils

import (
	"database/sql"

	e "github.com/techjanitor/pram-get/errors"
)

// user struct
type User struct {
	Id              uint   `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Group           uint   `json:"group"`
	IsConfirmed     bool   `json:"-"`
	IsLocked        bool   `json:"-"`
	IsBanned        bool   `json:"-"`
	IsAuthenticated bool   `json:"-"`
}

// get the user info from id
func (u *User) Info() (err error) {

	// this needs an id
	if u.Id == 0 || u.Id == 1 {
		return e.ErrInvalidParam
	}

	// get data from users table
	err = db.QueryRow("SELECT usergroup_id,user_name,user_email,user_confirmed,user_locked,user_banned FROM users WHERE user_id = ?", u.Id).Scan(&u.Group, &u.Name, &u.Email, &u.IsConfirmed, &u.IsLocked, &u.IsBanned)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return e.ErrInternalError
	}

	// if account is not confirmed
	if !u.IsConfirmed {
		return e.ErrUserNotConfirmed
	}

	// if locked
	if u.IsLocked {
		return e.ErrUserLocked
	}

	// if banned
	if u.IsBanned {
		return e.ErrUserBanned
	}

	// mark authenticated
	u.IsAuthenticated = true

	return

}
