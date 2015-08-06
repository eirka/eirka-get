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
	Term   string
	Result TagsType
}

// TagsType is the top level of the JSON response
type TagsType struct {
	Body []Tags `json:"tags"`
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

	// Get Database handle
	db, err := u.GetDb()
	if err != nil {
		return
	}

	tags := []Tags{}

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
	searchterm := fmt.Sprintf("%%%s%%", i.Term)

	rows, err := db.Query(`select count,tag_id,tag_name,tagtype_id
	FROM (select count(image_id) as count,ib_id,tags.tag_id,tag_name,tagtype_id
	FROM tags left join tagmap on tags.tag_id = tagmap.tag_id group by tag_id) as a 
	WHERE ib_id = ? AND tag_name LIKE ?
	ORDER BY count DESC`, i.Ib, searchterm)
	if err != nil {
		return err
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

	// Add pagedresponse to the response struct
	response.Body = tags

	// This is the data we will serialize
	i.Result = response

	return

}
