package models

import (
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

// FavoritedModel holds the parameters from the request and also the key for the cache
type FavoritedModel struct {
	Ib     uint
	Result FavoritedType
}

// FavoritedType is the top level of the JSON response
type FavoritedType struct {
	Body []OnlyImage `json:"favorited,omitempty"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *FavoritedModel) Get() (err error) {

	if i.Ib == 0 {
		return e.ErrNotFound
	}

	// Initialize response header
	response := FavoritedType{}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// SQL query to select the top 20 favorited images for a given image board (ib_id).
	// The query joins the favorites, images, posts, and threads tables to filter out deleted threads and posts.
	// It groups the results by image_id and orders them by the count of favorites in descending order.
	rows, err := dbase.Query(`
		SELECT image_id, image_file, image_thumbnail, image_tn_height, image_tn_width 
		FROM (
			SELECT favorites.image_id, image_file, image_thumbnail, image_tn_height, image_tn_width, COUNT(*) AS favorites
			FROM favorites
			INNER JOIN images ON favorites.image_id = images.image_id
			INNER JOIN posts ON images.post_id = posts.post_id
			INNER JOIN threads ON posts.thread_id = threads.thread_id
			WHERE ib_id = ? AND thread_deleted != 1 AND post_deleted != 1
			GROUP BY image_id
			ORDER BY favorites DESC
			LIMIT 20
		) AS favorited`, i.Ib)
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
			rows.Close() // Explicitly close rows before returning
			return err
		}
		// Append rows to info struct
		response.Body = append(response.Body, image)
	}
	if err = rows.Err(); err != nil {
		return err
	}

	// This is the data we will serialize
	i.Result = response

	return

}
