package models

import (
	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"

	u "github.com/eirka/eirka-get/utils"
)

// FavoritesModel holds the parameters from the request and also the key for the cache
type FavoritesModel struct {
	User   uint
	Ib     uint
	Page   uint
	Result FavoritesType
}

// FavoritesType is the top level of the JSON response
type FavoritesType struct {
	Body u.PagedResponse `json:"favorites"`
}

// FavoritesHeader is the header for favorites page
type FavoritesHeader struct {
	Images []OnlyImage `json:"images,omitempty"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *FavoritesModel) Get() (err error) {

	if i.Ib == 0 || i.User == 0 || i.Page == 0 {
		return e.ErrNotFound
	}

	// Initialize response header
	response := FavoritesType{}

	// Initialize struct for pagination
	paged := u.PagedResponse{}
	// Set current page to parameter
	paged.CurrentPage = i.Page
	// Set tags per index page to config setting
	paged.PerPage = config.Settings.Limits.PostsPerPage

	// Initialize struct for tag info
	tagheader := FavoritesHeader{}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// Get total favorites count and put it in pagination struct
	err = dbase.QueryRow(`SELECT count(*) FROM favorites
	INNER JOIN images on favorites.image_id = images.image_id
	INNER JOIN posts on images.post_id = posts.post_id
	INNER JOIN threads on posts.thread_id = threads.thread_id
	WHERE favorites.user_id = ? AND ib_id = ? AND thread_deleted != 1 AND post_deleted != 1`, i.User, i.Ib).Scan(&paged.Total)
	if err != nil {
		return
	}

	// Calculate Limit and total Pages
	paged.Get()

	// Return 404 if page requested is larger than actual pages
	if i.Page > paged.Pages {
		return e.ErrNotFound
	}

	// get images in users favorites with limits
	rows, err := dbase.Query(`SELECT images.image_id,image_file,image_thumbnail,image_tn_height,image_tn_width
	FROM favorites
	INNER JOIN images on favorites.image_id = images.image_id
	INNER JOIN posts on images.post_id = posts.post_id
	INNER JOIN threads on posts.thread_id = threads.thread_id
	WHERE favorites.user_id = ? AND ib_id = ? AND thread_deleted != 1 AND post_deleted != 1
	ORDER BY favorite_id DESC LIMIT ?,?`, i.User, i.Ib, paged.Limit, paged.PerPage)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		// Initialize posts struct
		image := OnlyImage{}
		// Scan rows and place column into struct
		err := rows.Scan(&image.ID, &image.File, &image.Thumb, &image.ThumbHeight, &image.ThumbWidth)
		if err != nil {
			return err
		}
		// Append rows to info struct
		tagheader.Images = append(tagheader.Images, image)
	}
	if rows.Err() != nil {
		return
	}

	// Add tags slice to items interface
	paged.Items = tagheader

	// Add pagedresponse to the response struct
	response.Body = paged

	// This is the data we will serialize
	i.Result = response

	return

}
