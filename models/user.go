package models

import (
	"database/sql"

	e "github.com/techjanitor/pram-get/errors"
	u "github.com/techjanitor/pram-get/utils"
)

// UserModel holds the parameters from the request and also the key for the cache
type UserModel struct {
	Id     uint
	Result UserType
}

// UserType is the top level of the JSON response
type UserType struct {
	User u.User `json:"user"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *UserModel) Get() (err error) {

	// Initialize response header
	response := UserType{}

	// Initialize struct for tag info
	user := u.User{}

	user.Id = i.Id

	// Get Database handle
	db, err := u.GetDb()
	if err != nil {
		return
	}

	// Get user name and group
	err = db.QueryRow("select user_name,usergroup_id from users where user_id = ?", i.Id).Scan(&user.Name, &user.Group)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return
	}

	// Add taginfo to the response struct
	response.User = user

	// This is the data we will serialize
	i.Result = response

	return

}
