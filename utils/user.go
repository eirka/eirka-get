package utils

import (
	"database/sql"

	e "github.com/techjanitor/pram-get/errors"
)

// user struct
type User struct {
	Id    uint   `json:"id"`
	Name  string `json:"name"`
	Group uint   `json:"group"`
}

// get the user info
func (u *User) Info() (err error) {

	err = db.QueryRow("SELECT usergroup_id,user_name FROM users WHERE user_id = ?", u.Id).Scan(&u.Group, &u.Name)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return
	}

	return

}
