package models

import (
	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"

	u "github.com/eirka/eirka-get/utils"
)

// DirectoryModel holds the parameters from the request and also the key for the cache
type DirectoryModel struct {
	Ib     uint
	Result DirectoryType
}

// IndexType is the top level of the JSON response
type DirectoryType struct {
	Body []Directory `json:"directory"`
}

// Thread directory
type Directory struct {
	Id     uint   `json:"id"`
	Title  string `json:"title"`
	Closed bool   `json:"closed"`
	Sticky bool   `json:"sticky"`
	Posts  uint   `json:"postcount"`
	Pages  uint   `json:"pages"`
	Last   string `json:"last_post"`
	Images uint   `json:"images"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *DirectoryModel) Get() (err error) {

	// Initialize response header
	response := DirectoryType{}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	rows, err := dbase.Query(`SELECT threads.thread_id,thread_title,thread_closed,thread_sticky,count(posts.post_id),count(image_id),thread_last_post 
	FROM threads
	LEFT JOIN posts on threads.thread_id = posts.thread_id
	LEFT JOIN images on images.post_id = posts.post_id
	WHERE ib_id = ? AND thread_deleted != 1 AND post_deleted != 1
	GROUP BY threads.thread_id
	ORDER BY thread_sticky = 1 DESC, thread_last_post DESC`, i.Ib)
	if err != nil {
		return
	}
	defer rows.Close()

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

	// Return 404 if there are no threads in ib
	if len(threads) == 0 {
		return e.ErrNotFound
	}

	// Add pagedresponse to the response struct
	response.Body = threads

	// This is the data we will serialize
	i.Result = response

	return

}
