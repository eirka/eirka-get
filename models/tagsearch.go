package models

import (
	"fmt"
	"strings"

	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/eirka/eirka-libs/validate"
)

// TagSearchModel holds the parameters from the request and also the key for the cache
type TagSearchModel struct {
	Ib     uint
	Term   string
	Result TagSearchType
}

// TagSearchType is the top level of the JSON response
type TagSearchType struct {
	Body []Tags `json:"tagsearch"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *TagSearchModel) Get() (err error) {

	// Initialize response header
	response := TagSearchType{}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	tags := []Tags{}

	// Validate tag input
	if i.Term != "" {
		tag := validate.Validate{Input: i.Term, Max: config.Settings.Limits.TagMaxLength, Min: config.Settings.Limits.TagMinLength}
		if tag.MinLength() {
			return e.ErrInvalidParam
		} else if tag.MaxLength() {
			return e.ErrInvalidParam
		}
	}

	// split search term
	terms := strings.Split(strings.TrimSpace(i.Term), " ")

	var searchterm string

	// add wildcards to the terms
	for i, term := range terms {
		// if not the first index then add a space before
		if i > 0 {
			searchterm += " "
		}
		// add a plus to the front of the term
		if len(term) > 0 && term != "" {
			searchterm += fmt.Sprintf("+%s", term)
		}
	}

	// add a wildcard to the end of the term
	searchterm += "*"

	rows, err := dbase.Query(`SELECT count,tag_id,tag_name,tagtype_id
	FROM (SELECT count(image_id) as count,ib_id,tags.tag_id,tag_name,tagtype_id
	FROM tags 
	LEFT JOIN tagmap on tags.tag_id = tagmap.tag_id 
	WHERE ib_id = ? AND MATCH(tag_name) AGAINST (? IN BOOLEAN MODE)
	group by tag_id) as a`, i.Ib, searchterm)
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

	// Add pagedresponse to the response struct
	response.Body = tags

	// This is the data we will serialize
	i.Result = response

	return

}
