package models

import (
	"database/sql"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

// FavoriteModel holds the parameters from the request and also the key for the cache
type FavoriteModel struct {
	User   uint
	ID     uint
	Result FavoriteType
}

// FavoriteType is the top level of the JSON response
type FavoriteType struct {
	Starred bool `json:"starred"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *FavoriteModel) Get() (err error) {

	if i.User == 0 || i.ID == 0 {
		return e.ErrNotFound
	}

	// Initialize response header
	response := FavoriteType{}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// See if a user has starred an image
	// The query checks if there is at least one row in the 'favorites' table
	// where the 'user_id' matches i.User and 'image_id' matches i.ID.
	// The 'EXISTS' function returns true if such a row exists, otherwise false.
	err = dbase.QueryRow(`
		SELECT EXISTS(
			SELECT 1 
			FROM favorites 
			WHERE user_id = ? 
			AND image_id = ?
		)`, i.User, i.ID).Scan(&response.Starred)
	if err != nil {
		if err == sql.ErrNoRows {
			// No rows means the image is not starred
			response.Starred = false
			err = nil
		} else {
			// Handle other database errors
			return err
		}
	}

	// This is the data we will serialize
	i.Result = response

	return

}
