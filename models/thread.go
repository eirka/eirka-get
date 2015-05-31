package models

import (
	"database/sql"

	"github.com/techjanitor/pram-get/config"
	e "github.com/techjanitor/pram-get/errors"
	u "github.com/techjanitor/pram-get/utils"
)

// ThreadModel holds the parameters from the request and also the key for the cache
type ThreadModel struct {
	Thread uint
	Page   uint
	Result ThreadType
}

// IndexType is the top level of the JSON response
type ThreadType struct {
	Body u.PagedResponse `json:"thread"`
}

// Info header for thread view
type ThreadInfo struct {
	Id     uint          `json:"id"`
	Title  string        `json:"title"`
	Closed bool          `json:"closed"`
	Sticky bool          `json:"sticky"`
	Posts  []ThreadPosts `json:"posts"`
}

// Thread Posts
type ThreadPosts struct {
	Id          uint    `json:"id"`
	Num         uint    `json:"num"`
	Name        *string `json:"name"`
	Time        *string `json:"time"`
	Text        *string `json:"comment"`
	ImgId       *uint   `json:"img_id,omitempty"`
	File        *string `json:"filename,omitempty"`
	Thumb       *string `json:"thumbnail,omitempty"`
	ThumbHeight *uint   `json:"tn_height,omitempty"`
	ThumbWidth  *uint   `json:"tn_width,omitempty"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *ThreadModel) Get() (err error) {

	// Initialize response header
	response := ThreadType{}

	// Initialize struct for pagination
	paged := u.PagedResponse{}
	// Set current page to parameter
	paged.CurrentPage = i.Page
	// Set threads per index page to config setting
	paged.PerPage = config.Settings.Limits.PostsPerPage

	// Initialize struct for thread info
	thread := ThreadInfo{}

	// Get Database handle
	db, err := u.GetDb()
	if err != nil {
		return
	}

	// Get total thread count and put it in pagination struct
	err = db.QueryRow(`SELECT threads.thread_id,thread_title,thread_closed,thread_sticky,count(posts.post_id) FROM threads 
	LEFT JOIN posts on threads.thread_id = posts.thread_id
	WHERE threads.thread_id = ?
	GROUP BY threads.thread_id`, i.Thread).Scan(&thread.Id, &thread.Title, &thread.Closed, &thread.Sticky, &paged.Total)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return
	}

	// Calculate Limit and total Pages
	paged.Get()

	// Check to see if thread has been deleted
	if u.GetBool("thread_deleted", "threads", "thread_id", i.Thread) {
		return e.ErrNotFound
	}

	// check page number
	switch i.Page {
	case 0:
		paged.PerPage = paged.Total
		paged.Limit = 0
	case i.Page > paged.Pages:
		return e.ErrNotFound
	}

	// Query rows
	rows, err := db.Query(`SELECT posts.post_id,post_num,post_name,post_time,post_text,image_id,image_file,image_thumbnail,image_tn_height,image_tn_width
	FROM posts
	LEFT JOIN images on posts.post_id = images.post_id
	WHERE posts.thread_id = ? ORDER BY post_id LIMIT ?,?`, i.Thread, paged.Limit, paged.PerPage)

	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		// Initialize posts struct
		post := ThreadPosts{}
		// Scan rows and place column into struct
		err := rows.Scan(&post.Id, &post.Num, &post.Name, &post.Time, &post.Text, &post.ImgId, &post.File, &post.Thumb, &post.ThumbHeight, &post.ThumbWidth)
		if err != nil {
			return err
		}
		// Append rows to info struct
		thread.Posts = append(thread.Posts, post)
	}
	err = rows.Err()
	if err != nil {
		return
	}

	// Add threads slice to items interface
	paged.Items = thread

	// Add pagedresponse to the response struct
	response.Body = paged

	// This is the data we will serialize
	i.Result = response

	return

}
