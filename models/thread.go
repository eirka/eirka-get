package models

import (
	"database/sql"
	"time"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"

	u "github.com/eirka/eirka-get/utils"
)

// ThreadModel holds the parameters from the request and also the key for the cache
type ThreadModel struct {
	Ib     uint
	Thread uint
	Page   uint
	Posts  uint
	Result ThreadType
}

// ThreadType is the top level of the JSON response
type ThreadType struct {
	Body u.PagedResponse `json:"thread"`
}

// ThreadInfo header for thread view
type ThreadInfo struct {
	ID     uint          `json:"id"`
	Title  string        `json:"title"`
	Closed bool          `json:"closed"`
	Sticky bool          `json:"sticky"`
	Posts  []ThreadPosts `json:"posts"`
}

// ThreadPosts holds the posts for a thread
type ThreadPosts struct {
	ID          uint       `json:"id"`
	Num         uint       `json:"num"`
	Name        string     `json:"name"`
	UID         uint       `json:"uid"`
	Group       uint       `json:"group"`
	Time        *time.Time `json:"time"`
	Text        *string    `json:"comment"`
	ImageID     *uint      `json:"img_id,omitempty"`
	File        *string    `json:"filename,omitempty"`
	Thumb       *string    `json:"thumbnail,omitempty"`
	ThumbHeight *uint      `json:"tn_height,omitempty"`
	ThumbWidth  *uint      `json:"tn_width,omitempty"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *ThreadModel) Get() (err error) {

	if i.Ib == 0 || i.Thread == 0 {
		return e.ErrNotFound
	}

	// Initialize response header
	response := ThreadType{}

	// Initialize struct for pagination
	paged := u.PagedResponse{}
	// Set current page to parameter
	paged.CurrentPage = i.Page
	// Set threads per index page to config setting
	paged.PerPage = i.Posts

	// Initialize struct for thread info
	thread := ThreadInfo{}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// This SQL query retrieves thread information and post count for a specific thread.
	// It joins the threads and posts tables, filtering by thread_id and ib_id.
	// The query ensures that only non-deleted threads and posts are counted.
	// It returns the thread ID, title, closed status, sticky status, and total post count.
	err = dbase.QueryRow(`
        SELECT 
            threads.thread_id, thread_title, thread_closed, thread_sticky, COUNT(posts.post_id)
        FROM 
            threads
        INNER JOIN 
            posts ON threads.thread_id = posts.thread_id
        WHERE 
            threads.thread_id = ? 
            AND threads.ib_id = ? 
            AND thread_deleted != 1 
            AND post_deleted != 1
        GROUP BY 
            threads.thread_id
    `, i.Thread, i.Ib).Scan(&thread.ID, &thread.Title, &thread.Closed, &thread.Sticky, &paged.Total)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return
	}

	// Calculate Limit and total Pages
	paged.Get()

	// check page number
	switch {
	case i.Page == 0:
		paged.PerPage = paged.Total
		paged.Limit = 0
	case i.Page > paged.Pages:
		return e.ErrNotFound
	}

	// This SQL query retrieves detailed information about posts in a specific thread.
	// It joins multiple tables (posts, images, users, user_role_map) to gather all necessary data.
	// The query uses a COALESCE function to determine the user's role, considering both global and image board-specific roles.
	// It filters for non-deleted posts, orders them by post_id, and applies pagination using LIMIT.
	rows, err := dbase.Query(`
        SELECT 
            posts.post_id, post_num, user_name, users.user_id,
            COALESCE(
                (SELECT MAX(role_id) 
                 FROM user_ib_role_map 
                 WHERE user_ib_role_map.user_id = users.user_id AND ib_id = ?),
                user_role_map.role_id
            ) AS role,
            post_time, post_text, image_id, image_file, image_thumbnail, image_tn_height, image_tn_width
        FROM 
            posts
        LEFT JOIN 
            images ON posts.post_id = images.post_id
        INNER JOIN 
            users ON posts.user_id = users.user_id
        INNER JOIN 
            user_role_map ON (user_role_map.user_id = users.user_id)
        WHERE 
            posts.thread_id = ? 
            AND post_deleted != 1
        ORDER BY 
            post_id 
        LIMIT ?, ?
    `, i.Ib, i.Thread, paged.Limit, paged.PerPage)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		// Initialize posts struct
		post := ThreadPosts{}
		// Scan rows and place column into struct
		err := rows.Scan(&post.ID, &post.Num, &post.Name, &post.UID, &post.Group, &post.Time, &post.Text, &post.ImageID, &post.File, &post.Thumb, &post.ThumbHeight, &post.ThumbWidth)
		if err != nil {
			rows.Close() // Explicitly close rows before returning
			return err
		}
		// Append rows to info struct
		thread.Posts = append(thread.Posts, post)
	}
	if err = rows.Err(); err != nil {
		return err
	}

	// Add threads slice to items interface
	paged.Items = thread

	// Add pagedresponse to the response struct
	response.Body = paged

	// This is the data we will serialize
	i.Result = response

	return

}
