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

	title := validate.Validate{Input: i.Term, Max: config.Settings.Limits.TitleMaxLength, Min: config.Settings.Limits.TitleMinLength}
	if title.IsEmpty() {
		return e.ErrNoTitle
	} else if title.MinLength() {
		return e.ErrTitleShort
	} else if title.MaxLength() {
		return e.ErrTitleLong
	}

	terms := strings.Split(strings.TrimSpace(i.Term), " ")

	var searchquery []string

	for _, term := range terms {
		// get rid of bad cahracters for mysql in boolean mode
		term = formatQuery(term)

		searchquery = append(searchquery, fmt.Sprintf("+%s", term))
	}

	// This SQL query performs a full-text search on thread titles and retrieves relevant thread information.
	// It joins the threads, posts, and images tables to gather comprehensive data about each matching thread.
	// The query does the following:
	// 1. Selects thread details, including ID, title, closed/sticky status, post count, and image count
	// 2. Calculates the last post time for each thread (excluding deleted posts)
	// 3. Filters results by image board ID and excludes deleted threads/posts
	// 4. Uses MATCH...AGAINST for full-text search on thread titles with boolean mode and wildcard
	// 5. Groups results by thread ID to avoid duplicates
	// 6. Orders results by the last post time (most recent first)
	rows, err := dbase.Query(`
        SELECT 
            threads.thread_id,
            thread_title,
            thread_closed,
            thread_sticky,
            COUNT(posts.post_id),
            COUNT(image_id),
            (SELECT MAX(post_time) 
             FROM posts 
             WHERE thread_id = threads.thread_id AND post_deleted != 1) AS thread_last_post
        FROM threads
        LEFT JOIN posts ON threads.thread_id = posts.thread_id
        LEFT JOIN images ON images.post_id = posts.post_id
        WHERE ib_id = ? 
          AND thread_deleted != 1 
          AND post_deleted != 1
          AND MATCH(thread_title) AGAINST (? IN BOOLEAN MODE)
        GROUP BY threads.thread_id
        ORDER BY thread_last_post DESC
    `, i.Ib, fmt.Sprintf("%s*", strings.Join(searchquery, " ")))
	if err != nil {
		return
	}

	threads := []Directory{}
	for rows.Next() {
		thread := Directory{}
		err := rows.Scan(&thread.ID, &thread.Title, &thread.Closed, &thread.Sticky, &thread.Posts, &thread.Images, &thread.Last)
		if err != nil {
			rows.Close() // Explicitly close rows before returning
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
	if err = rows.Err(); err != nil {
		return err
	}

	// Add pagedresponse to the response struct
	response.Body = threads

	// This is the data we will serialize
	i.Result = response

	return

}
