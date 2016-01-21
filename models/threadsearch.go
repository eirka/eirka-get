package models

import (
	"fmt"
	"strings"

	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/eirka/eirka-libs/validate"

	u "github.com/eirka/eirka-get/utils"
)

// ThreadSearchModel holds the parameters from the request and also the key for the cache
type ThreadSearchModel struct {
	Ib     uint
	Term   string
	Result ThreadSearchType
}

// ThreadSearchType is the top level of the JSON response
type ThreadSearchType struct {
	Body []Directory `json:"threadsearch"`
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
    AND MATCH(thread_title) AGAINST (? IN BOOLEAN MODE)
    GROUP BY threads.thread_id
    ORDER BY thread_last_post`, i.Ib, wildterm)
	if err != nil {
		return
	}

	threads := []Directory{}
	for rows.Next() {
		thread := Directory{}
		err := rows.Scan(&thread.Id, &thread.Title, &thread.Closed, &thread.Sticky, &thread.Posts, &thread.Images, &thread.Last)
		if err != nil {
			return err
		}

		// Get the number of pages in the thread
		postpages := u.PagedResponse{}
		postpages.Total = thread.Posts
		postpages.CurrentPage = 1
		postpages.PerPage = config.Settings.Limits.PostsPerPage
		postpages.Get()

		// set pages
		thread.Pages = postpages.Pages

		threads = append(threads, thread)
	}
	err = rows.Err()
	if err != nil {
		return
	}

	// Add pagedresponse to the response struct
	response.Body = threads

	// This is the data we will serialize
	i.Result = response

	return

}
