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

	rows, err := dbase.Query(`SELECT request_itemvalue,image_file,image_thumbnail,image_tn_height,image_tn_width FROM
    (SELECT request_itemvalue,image_file,image_thumbnail,image_tn_height,image_tn_width,COUNT(request_itemvalue) AS hits
    FROM analytics
    INNER JOIN images on request_itemvalue = images.image_id
	INNER JOIN posts on images.post_id = posts.post_id
	INNER JOIN threads on posts.thread_id = threads.thread_id
    WHERE analytics.ib_id = ? AND request_itemkey = "image" AND request_time >= (now() - interval 3 day)
    AND thread_deleted != 1 AND post_deleted != 1
    GROUP BY request_itemvalue
    ORDER BY hits DESC LIMIT 50) AS popular`, i.Ib)
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
		response.Body = append(response.Body, image)
	}
	if rows.Err() != nil {
		return
	}

	// This is the data we will serialize
	i.Result = response

	return

}
