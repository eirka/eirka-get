package models

import (
	"github.com/techjanitor/pram-get/config"
	e "github.com/techjanitor/pram-get/errors"
	u "github.com/techjanitor/pram-get/utils"
)

// IndexModel holds the parameters from the request and also the key for the cache
type IndexModel struct {
	Ib     uint
	Page   uint
	Result IndexType
}

// ThreadIds holds all the thread ids for the loop that gets the posts
type ThreadIds struct {
	Id       uint
	Title    string
	LastPost string
	Closed   bool
	Sticky   bool
	Total    uint
}

// IndexType is the top level of the JSON response
type IndexType struct {
	Body u.PagedResponse `json:"index"`
}

// IndexThreadHeader holds the information for the threads
type IndexThreadHeader struct {
	Id        uint          `json:"id"`
	Title     string        `json:"title"`
	Closed    bool          `json:"closed"`
	Sticky    bool          `json:"sticky"`
	OmitPosts uint          `json:"omit_posts"`
	Pages     uint          `json:"last_page"`
	Posts     []ThreadPosts `json:"posts"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *IndexModel) Get() (err error) {

	// Initialize response header
	response := IndexType{}

	// Initialize struct for pagination
	paged := u.PagedResponse{}
	// Set current page to parameter
	paged.CurrentPage = i.Page
	// Set threads per index page to config setting
	paged.PerPage = config.Settings.Limits.ThreadsPerPage

	// Initialize struct for all thread ids
	thread_ids := []ThreadIds{}

	// Get Database handle
	db, err := u.GetDb()
	if err != nil {
		return
	}

	// Get total thread count and put it in pagination struct
	err = db.QueryRow("select count(*) from threads where ib_id = ?", i.Ib).Scan(&paged.Total)
	if err != nil {
		return
	}

	// Calculate Limit and total Pages
	paged.Get()

	// Return 404 if page requested is larger than actual pages
	if i.Page > paged.Pages {
		return e.ErrNotFound
	}

	// Get all thread ids with limit
	thread_id_rows, err := db.Query(`SELECT threads.thread_id,thread_title,thread_closed,thread_sticky,count(posts.post_id)
	FROM threads
	LEFT JOIN posts on threads.thread_id = posts.thread_id
	WHERE ib_id = ? AND thread_deleted = 0
	GROUP BY threads.thread_id
	ORDER BY thread_sticky = 1 DESC, thread_last_post DESC LIMIT ?,?`, i.Ib, paged.Limit, config.Settings.Limits.ThreadsPerPage)
	if err != nil {
		return
	}
	defer thread_id_rows.Close()

	for thread_id_rows.Next() {
		// Initialize posts struct
		thread_id_row := ThreadIds{}
		// Scan rows and place column into struct
		err := thread_id_rows.Scan(&thread_id_row.Id, &thread_id_row.Title, &thread_id_row.Closed, &thread_id_row.Sticky, &thread_id_row.Total)
		if err != nil {
			return err
		}
		// Append rows to info struct
		thread_ids = append(thread_ids, thread_id_row)
	}
	err = thread_id_rows.Err()
	if err != nil {
		return
	}

	//Get last thread posts
	ps1, err := db.Prepare(`SELECT * FROM
        (SELECT posts.post_id,post_num,post_name,post_time,post_text,image_id,image_file,image_thumbnail,image_tn_height,image_tn_width FROM posts
        LEFT JOIN images on posts.post_id = images.post_id
        WHERE posts.thread_id = ? ORDER BY post_num = 1 DESC, post_num DESC LIMIT ?)
        AS p ORDER BY post_num ASC`)
	if err != nil {
		return
	}
	defer ps1.Close()

	// Initialize slice for threads
	threads := []IndexThreadHeader{}

	// Loop over the values of thread_ids
	for _, id := range thread_ids {

		thread := IndexThreadHeader{}

		// Get last page from thread
		postpages := u.PagedResponse{}
		postpages.Total = id.Total
		postpages.CurrentPage = 1
		postpages.PerPage = config.Settings.Limits.PostsPerPage
		postpages.Get()

		// Set thread fields
		thread.Id = id.Id
		thread.Title = id.Title
		thread.Closed = id.Closed
		thread.Sticky = id.Sticky
		thread.Pages = postpages.Pages

		// Get omitted postcount
		if id.Total <= config.Settings.Limits.PostsPerThread {
			thread.OmitPosts = 0
		} else {
			thread.OmitPosts = (id.Total - config.Settings.Limits.PostsPerThread)
		}

		e1, err := ps1.Query(id.Id, config.Settings.Limits.PostsPerThread)
		if err != nil {
			return err
		}
		defer e1.Close()

		for e1.Next() {
			// Initialize posts struct
			post := ThreadPosts{}
			// Scan rows and place column into struct
			err := e1.Scan(&post.Id, &post.Num, &post.Name, &post.Time, &post.Text, &post.ImgId, &post.File, &post.Thumb, &post.ThumbHeight, &post.ThumbWidth)
			if err != nil {
				return err
			}
			// Append rows to info struct
			thread.Posts = append(thread.Posts, post)
		}
		err = e1.Err()
		if err != nil {
			return err
		}
		threads = append(threads, thread)
	}

	// Add threads slice to items interface
	paged.Items = threads

	// Add pagedresponse to the response struct
	response.Body = paged

	// This is the data we will serialize
	i.Result = response

	return

}
