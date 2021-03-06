package models

import (
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

	// see if a user has starred an image
	err = dbase.QueryRow("select count(*) from favorites where user_id = ? AND image_id = ? LIMIT 1", i.User, i.ID).Scan(&response.Starred)
	if err != nil {
		return
	}

	// This is the data we will serialize
	i.Result = response

	return

}
