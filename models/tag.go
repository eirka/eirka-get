package models

import (
	"database/sql"

	"github.com/techjanitor/pram-get/config"
	e "github.com/techjanitor/pram-get/errors"
	u "github.com/techjanitor/pram-get/utils"
)

// TagModel holds the parameters from the request and also the key for the cache
type TagModel struct {
	Tag    uint
	Page   uint
	Result TagType
}

// IndexType is the top level of the JSON response
type TagType struct {
	Body u.PagedResponse `json:"tag"`
}

// Header for tag page
type TagHeader struct {
	Id     uint       `json:"id"`
	Tag    *string    `json:"tag"`
	Type   *uint      `json:"type"`
	Images []TagImage `json:"images,omitempty"`
}

// Image struct for tag page
type TagImage struct {
	Id          uint    `json:"id"`
	Thumb       *string `json:"thumbnail"`
	ThumbHeight *uint   `json:"tn_height"`
	ThumbWidth  *uint   `json:"tn_width"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *TagModel) Get() (err error) {

	// Initialize response header
	response := TagType{}

	// Initialize struct for pagination
	paged := u.PagedResponse{}
	// Set current page to parameter
	paged.CurrentPage = i.Page
	// Set tags per index page to config setting
	paged.PerPage = config.Settings.Limits.PostsPerPage

	// Initialize struct for tag info
	tagheader := TagHeader{}

	tagheader.Id = i.Tag

	// Get Database handle
	db, err := u.GetDb()
	if err != nil {
		return
	}

	// Get tag name and type
	err = db.QueryRow("select tag_name,tagtype_id from tags where tag_id = ?", i.Tag).Scan(&tagheader.Tag, &tagheader.Type)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return
	}

	// Get total tag count
	// Get total tag count and put it in pagination struct
	err = db.QueryRow(`select count(*) from tagmap where tag_id = ?`, i.Tag).Scan(&paged.Total)
	if err != nil {
		return
	}

	// Calculate Limit and total Pages
	paged.Get()

	// Return 404 if page requested is larger than actual pages
	if i.Page > paged.Pages {
		return e.ErrNotFound
	}

	// Set perpage and limit to total and 0 if page num is 0
	if i.Page == 0 {
		paged.PerPage = paged.Total
		paged.Limit = 0
	}

	rows, err := db.Query(`SELECT images.image_id,image_thumbnail,image_tn_height,image_tn_width 
	FROM tagmap
	LEFT JOIN images on tagmap.image_id = images.image_id
	WHERE tagmap.tag_id = ? ORDER BY tagmap.image_id LIMIT ?,?`, i.Tag, paged.Limit, paged.PerPage)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		// Initialize posts struct
		image := TagImage{}
		// Scan rows and place column into struct
		err := rows.Scan(&image.Id, &image.Thumb, &image.ThumbHeight, &image.ThumbWidth)
		if err != nil {
			return err
		}
		// Append rows to info struct
		tagheader.Images = append(tagheader.Images, image)
	}
	err = rows.Err()
	if err != nil {
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
