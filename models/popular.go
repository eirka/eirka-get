package models

import (
	u "github.com/techjanitor/pram-get/utils"
)

// PopularModel holds the parameters from the request and also the key for the cache
type PopularModel struct {
	Ib     uint
	Result PopularType
}

// PopularType is the top level of the JSON response
type PopularType struct {
	Body []PopularImage `json:"popular,omitempty"`
}

// Image struct for tag page
type PopularImage struct {
	Id          uint    `json:"id"`
	Thumb       *string `json:"thumbnail"`
	ThumbHeight *uint   `json:"tn_height"`
	ThumbWidth  *uint   `json:"tn_width"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *PopularModel) Get() (err error) {

	// Initialize response header
	response := PopularType{}

	// Get Database handle
	db, err := u.GetDb()
	if err != nil {
		return
	}

	rows, err := db.Query(`SELECT request_itemvalue,image_thumbnail,image_tn_height,image_tn_width FROM
    (SELECT request_itemvalue,image_thumbnail,image_tn_height,image_tn_width, COUNT(*) AS hits
    FROM analytics
    INNER JOIN images on request_itemvalue = images.image_id
    WHERE request_itemkey = "image" AND request_time > (now() - interval 3 day) AND ib_id = ?
    GROUP BY request_itemvalue
    ORDER BY hits DESC LIMIT 50) AS popular`, i.Ib)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		// Initialize posts struct
		image := PopularImage{}
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
