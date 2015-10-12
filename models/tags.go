package models

import (
	"fmt"

	"github.com/techjanitor/pram-get/config"
	e "github.com/techjanitor/pram-get/errors"
	u "github.com/techjanitor/pram-get/utils"
)

// TagsModel holds the parameters from the request and also the key for the cache
type TagsModel struct {
	Ib     uint
	Page   uint
	Term   string
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
	db, err := u.GetDb()
	if err != nil {
		return
	}

	// Validate tag input
	if i.Term != "" {
		tag := u.Validate{Input: i.Term, Max: config.Settings.Limits.TagMaxLength, Min: config.Settings.Limits.TagMinLength}
		if tag.MinLength() {
			return e.ErrInvalidParam
		} else if tag.MaxLength() {
			return e.ErrInvalidParam
		}
	}

	// add wildcards to the term
	searchterm := fmt.Sprintf("%s%%", i.Term)

	rows, err := db.Query(`SELECT count,tag_id,tag_name,tagtype_id
	FROM (SELECT count(image_id) as count,ib_id,tags.tag_id,tag_name,tagtype_id
	FROM tags 
	LEFT JOIN tagmap on tags.tag_id = tagmap.tag_id 
	WHERE ib_id = ? AND tag_name LIKE ?
	group by tag_id) as a 
	ORDER BY tag_name ASC`, i.Ib, searchterm)
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

	// Return 404 if there are no threads in ib
	if len(tags) == 0 {
		return e.ErrNotFound
	}

	paged.Total = len(tags)

	// Calculate Limit and total Pages
	paged.Get()

	// Add threads slice to items interface
	paged.Items = tags

	// Add pagedresponse to the response struct
	response.Body = paged

	// This is the data we will serialize
	i.Result = response

	return

}
