package models

import (
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

// This function has been moved to utils.FormatQuery

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

	// Build query and parameters safely with each term as a separate parameter
	var params []interface{}
	var booleanPlaceholders []string

	// First parameter is the image board ID
	params = append(params, i.Ib)

	// Build the boolean mode search string with placeholders
	for _, term := range terms {
		// Clean term for MySQL boolean mode
		term = u.FormatQuery(term)
		if term == "" {
			continue // Skip empty terms
		}

		// For boolean search with wildcard
		booleanPlaceholders = append(booleanPlaceholders, "+?*")
		params = append(params, term)
	}

	// Construct the search expression for the WHERE clause
	var booleanWhereExpr string
	if len(booleanPlaceholders) == 0 {
		booleanWhereExpr = "MATCH(thread_title) AGAINST ('' IN BOOLEAN MODE)"
	} else {
		booleanWhereExpr = "MATCH(thread_title) AGAINST (CONCAT(" + strings.Join(booleanPlaceholders, ", ' ', ") + ") IN BOOLEAN MODE)"
	}

	// This SQL query performs a full-text search on thread titles and retrieves relevant thread information.
	// Now using proper parameterization for security:
	// 1. Selects thread details, including ID, title, closed/sticky status, post count, and image count
	// 2. Calculates the last post time for each thread (excluding deleted posts)
	// 3. Filters results by image board ID and excludes deleted threads/posts
	// 4. Uses MATCH...AGAINST for full-text search on thread titles with parameterized inputs
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
          AND `+booleanWhereExpr+`
        GROUP BY threads.thread_id
        ORDER BY thread_last_post DESC
        LIMIT 100
    `, params...)
	if err != nil {
		return
	}
	// Ensure rows are closed even if there's an error later in the function
	defer rows.Close()

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
