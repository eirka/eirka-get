package models

import (
	u "github.com/techjanitor/pram-get/utils"
)

// FavoritedModel holds the parameters from the request and also the key for the cache
type FavoritedModel struct {
	Ib     uint
	Result FavoritedType
}

// FavoritedType is the top level of the JSON response
type FavoritedType struct {
	Body []FavoritedImage `json:"favorited,omitempty"`
}

// Image struct for tag page
type FavoritedImage struct {
	Id          uint    `json:"id"`
	Thumb       *string `json:"thumbnail"`
	ThumbHeight *uint   `json:"tn_height"`
	ThumbWidth  *uint   `json:"tn_width"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *FavoritedModel) Get() (err error) {

	// Initialize response header
	response := FavoritedType{}

	// Get Database handle
	db, err := u.GetDb()
	if err != nil {
		return
	}

	rows, err := db.Query(`SELECT image_id,image_thumbnail,image_tn_height,image_tn_width FROM
    (SELECT favorites.image_id,image_thumbnail,image_tn_height,image_tn_width,COUNT(*) AS favorites
    FROM favorites
    INNER JOIN images ON favorites.image_id = images.image_id
 	INNER JOIN posts on images.post_id = posts.post_id 
	INNER JOIN threads on posts.thread_id = threads.thread_id 
	WHERE ib_id = ? AND thread_deleted != 1 AND post_deleted != 1
    GROUP BY image_id
    ORDER BY favorites DESC
    LIMIT 20) AS favorited`, i.Ib)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		// Initialize posts struct
		image := FavoritedImage{}
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
