package models

import (
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

// PopularModel holds the parameters from the request and also the key for the cache
type PopularModel struct {
	Ib     uint
	Result PopularType
}

// PopularType is the top level of the JSON response
type PopularType struct {
	Body []OnlyImage `json:"popular,omitempty"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *PopularModel) Get() (err error) {

	if i.Ib == 0 {
		return e.ErrNotFound
	}

	// Initialize response header
	response := PopularType{}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// SQL query to select the most popular images based on the number of hits in the last 3 days.
	// It joins the analytics, images, posts, and threads tables to gather the necessary data.
	// The results are filtered to exclude deleted threads and posts, and are limited to the top 50 hits.
	rows, err := dbase.Query(`
		SELECT request_itemvalue, image_file, image_thumbnail, image_tn_height, image_tn_width 
		FROM (
			SELECT request_itemvalue, image_file, image_thumbnail, image_tn_height, image_tn_width, COUNT(request_itemvalue) AS hits
			FROM analytics
			INNER JOIN images ON request_itemvalue = images.image_id
			INNER JOIN posts ON images.post_id = posts.post_id
			INNER JOIN threads ON posts.thread_id = threads.thread_id
			WHERE analytics.ib_id = ? 
			AND request_itemkey = "image" 
			AND request_time >= (NOW() - INTERVAL 3 DAY)
			AND thread_deleted != 1 
			AND post_deleted != 1
			GROUP BY request_itemvalue
			ORDER BY hits DESC 
			LIMIT 50
		) AS popular`, i.Ib)
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
