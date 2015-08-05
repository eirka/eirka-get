package models

import (
	u "github.com/techjanitor/pram-get/utils"
)

// NewModel holds the parameters from the request and also the key for the cache
type NewModel struct {
	Ib     uint
	Result NewType
}

// NewType is the top level of the JSON response
type NewType struct {
	Body []NewImage `json:"new,omitempty"`
}

// Image struct for tag page
type NewImage struct {
	Id          uint    `json:"id"`
	Thumb       *string `json:"thumbnail"`
	ThumbHeight *uint   `json:"tn_height"`
	ThumbWidth  *uint   `json:"tn_width"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *NewModel) Get() (err error) {

	// Initialize response header
	response := NewType{}

	// Get Database handle
	db, err := u.GetDb()
	if err != nil {
		return
	}

	rows, err := db.Query(`SELECT images.image_id,image_thumbnail,image_tn_height,image_tn_width 
	FROM images
	LEFT JOIN posts on images.post_id = posts.post_id 
	LEFT JOIN threads on posts.thread_id = threads.thread_id 
	WHERE ib_id = ? ORDER BY images.image_id DESC LIMIT 20`, i.Ib)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		// Initialize posts struct
		image := NewImage{}
		// Scan rows and place column into struct
		err := rows.Scan(&image.Id, &image.Thumb, &image.ThumbHeight, &image.ThumbWidth)
		if err != nil {
			return err
		}
		// Append rows to info struct
		response.Body = append(response.Body, image)
	}
	err = rows.Err()
	if err != nil {
		return
	}

	// This is the data we will serialize
	i.Result = response

	return

}
