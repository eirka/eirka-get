package models

import (
	"github.com/techjanitor/pram-libs/db"
	e "github.com/techjanitor/pram-libs/errors"
)

// UserModel holds the parameters from the request and also the key for the cache
type UserModel struct {
	User   uint
	Ib     uint
	Result UserType
}

// UserType is the top level of the JSON response
type UserType struct {
	User UserInfo `json:"user"`
}

// UserInfo holds all the user metadata
type UserInfo struct {
	Id     uint   `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Group  uint   `json:"group"`
	Avatar string `json:"avatar"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *UserModel) Get() (err error) {

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
    user_name,user_email,user_avatar FROM users
    INNER JOIN user_role_map ON (user_role_map.user_id = users.user_id)
    WHERE users.user_id = ?`, i.Ib, i.User).Scan(&r.Group, &r.Name, &r.Email, &r.Avatar)
	if err != nil {
		return
	}

	// This is the data we will serialize
	i.Result = response

	return

}
