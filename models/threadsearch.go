package models

import (
	"fmt"
	"strings"

	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/eirka/eirka-libs/validate"
)

// ThreadSearchModel holds the parameters from the request and also the key for the cache
type ThreadSearchModel struct {
	Ib     uint
	Term   string
	Result ThreadSearchType
}

// ThreadSearchType is the top level of the JSON response
type ThreadSearchType struct {
	Body []Tags `json:"ThreadSearch"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *ThreadSearchModel) Get() (err error) {

	// Initialize response header
	response := ThreadSearchType{}

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

	// add plusses to the terms
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
	wildterm := fmt.Sprintf("%s*", searchterm)

	rows, err := dbase.Query(`SELECT threads.thread_id,thread_title,thread_closed,thread_sticky,count(posts.post_id),count(image_id),thread_last_post 
    FROM threads
    LEFT JOIN posts on threads.thread_id = posts.thread_id
    LEFT JOIN images on images.post_id = posts.post_id
    WHERE ib_id = ? AND thread_deleted != 1 AND post_deleted != 1
    AND MATCH(tag_name) AGAINST (? IN BOOLEAN MODE)
    GROUP BY threads.thread_id
    ORDER BY thread_last_post`, searchterm, i.Ib, wildterm)
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
