package utils

import (
	"database/sql"

	e "github.com/techjanitor/pram-get/errors"
)

// user struct
type User struct {
	Id        uint   `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Group     uint   `json:"group"`
	Confirmed bool   `json:"-"`
	Locked    bool   `json:"-"`
	Banned    bool   `json:"-"`
}

// get the user info from id
func (u *User) Info() (err error) {

	// this needs an id
	if u.Id == 0 {
		return e.ErrInvalidParam
	}

	err = db.QueryRow("SELECT usergroup_id,user_name,user_email,user_confirmed,user_locked,user_banned FROM users WHERE user_id = ?", u.Id).Scan(&u.Group, &u.Name, &u.Email, &u.Confirmed, &u.Locked, &u.Banned)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return
	}

	return

}
