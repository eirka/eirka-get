package models

import (
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

// NewModel holds the parameters from the request and also the key for the cache
type NewModel struct {
	Ib     uint
	Result NewType
}

// NewType is the top level of the JSON response
type NewType struct {
	Body []OnlyImage `json:"new,omitempty"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *NewModel) Get() (err error) {

	if i.Ib == 0 {
		return e.ErrNotFound
	}

	// Initialize response header
	response := NewType{}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// SQL query to select image details from the database.
	// The query joins the images, posts, and threads tables to retrieve image information
	// where the image board ID matches the provided ID, and the thread and post are not deleted.
	// The results are ordered by image ID in descending order and limited to 20 records.
	rows, err := dbase.Query(`
		SELECT 
			images.image_id, 
			image_file, 
			image_thumbnail, 
			image_tn_height, 
			image_tn_width
		FROM 
			images
		INNER JOIN 
			posts ON images.post_id = posts.post_id
		INNER JOIN 
			threads ON posts.thread_id = threads.thread_id
		WHERE 
			ib_id = ? 
			AND thread_deleted != 1 
			AND post_deleted != 1
		ORDER BY 
			images.image_id DESC 
		LIMIT 20`, i.Ib)
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
