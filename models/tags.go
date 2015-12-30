package models

import (
	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"

	u "github.com/eirka/eirka-get/utils"
)

// TagsModel holds the parameters from the request and also the key for the cache
type TagsModel struct {
	Ib     uint
	Page   uint
	Result TagsType
}

// TagsType is the top level of the JSON response
type TagsType struct {
	Body u.PagedResponse `json:"tags"`
}

// Taglist struct
type Tags struct {
	Id    uint   `json:"id"`
	Tag   string `json:"tag"`
	Total uint   `json:"total"`
	Type  uint   `json:"type"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *TagsModel) Get() (err error) {

	// Initialize response header
	response := TagsType{}

	// tags slice
	tags := []Tags{}

	// Initialize struct for pagination
	paged := u.PagedResponse{}
	// Set current page to parameter
	paged.CurrentPage = i.Page
	// Set threads per index page to config setting
	paged.PerPage = config.Settings.Limits.PostsPerPage

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// Get total tag count and put it in pagination struct
	err = dbase.QueryRow("SELECT count(*) FROM tags WHERE ib_id = ?", i.Ib).Scan(&paged.Total)
	if err != nil {
		return
	}

	// Calculate Limit and total Pages
	paged.Get()

	// check page number
	switch {
	case i.Page == 0:
		paged.PerPage = paged.Total
		paged.Limit = 0
	case i.Page > paged.Pages:
		return e.ErrNotFound
	}

	// get image counts from tagmap
	rows, err := dbase.Query(`SELECT * FROM 
    (SELECT count(tagmap.image_id) as count,tags.tag_id,tag_name,tagtype_id FROM tags
    LEFT JOIN tagmap on tags.tag_id = tagmap.tag_id 
    LEFT JOIN images on tagmap.image_id = images.image_id
    LEFT JOIN posts on images.post_id = posts.post_id 
    LEFT JOIN threads on posts.thread_id = threads.thread_id 
    WHERE tags.ib_id = ? AND (thread_deleted != 1 OR thread_deleted IS NULL) AND (post_deleted != 1 OR post_deleted IS NULL)
    GROUP by tag_id) as a
    ORDER BY count DESC LIMIT ?,?`, i.Ib, paged.Limit, paged.PerPage)
	if err != nil {
		return
	}

	for rows.Next() {
		// Initialize posts struct
		tag := Tags{}
		// Scan rows and place column into struct
		err := rows.Scan(&tag.Total, &tag.Id, &tag.Tag, &tag.Type)
		if err != nil {
			return err
		}

		// Append rows to info struct
		tags = append(tags, tag)
	}
	err = rows.Err()
	if err != nil {
		return
	}

	// Add threads slice to items interface
	paged.Items = tags

	// Add pagedresponse to the response struct
	response.Body = paged

	// This is the data we will serialize
	i.Result = response

	return

}
