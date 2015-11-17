package models

import (
	"database/sql"

	"github.com/techjanitor/pram-libs/config"
	"github.com/techjanitor/pram-libs/db"
	e "github.com/techjanitor/pram-libs/errors"

	u "github.com/techjanitor/pram-get/utils"
)

// TagModel holds the parameters from the request and also the key for the cache
type TagModel struct {
	Ib     uint
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
	Id     uint        `json:"id"`
	Tag    *string     `json:"tag"`
	Type   *uint       `json:"type"`
	Images []OnlyImage `json:"images,omitempty"`
}

// Image struct for tag page
type OnlyImage struct {
	Id          uint    `json:"id"`
	File        *string `json:"filename"`
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
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// Get tag name and type
	err = dbase.QueryRow(`SELECT tag_name, tagtype_id, count(image_id) FROM tags 
    INNER JOIN tagmap on tags.tag_id = tagmap.tag_id 
    WHERE tags.tag_id = ? AND ib_id = ?`, i.Tag, i.Ib).Scan(&tagheader.Tag, &tagheader.Type, &paged.Total)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
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

	rows, err := dbase.Query(`SELECT images.image_id,image_file,image_thumbnail,image_tn_height,image_tn_width 
	FROM tagmap
	INNER JOIN images on tagmap.image_id = images.image_id
    INNER JOIN posts on images.post_id = posts.post_id 
    INNER JOIN threads on posts.thread_id = threads.thread_id 
	WHERE tagmap.tag_id = ? AND thread_deleted != 1 AND post_deleted != 1
	ORDER BY tagmap.image_id LIMIT ?,?`, i.Tag, paged.Limit, paged.PerPage)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		// Initialize posts struct
		image := OnlyImage{}
		// Scan rows and place column into struct
		err := rows.Scan(&image.Id, &image.File, &image.Thumb, &image.ThumbHeight, &image.ThumbWidth)
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
