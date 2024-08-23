package models

import (
	"database/sql"
	"time"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

// PostModel holds the parameters from the request and also the key for the cache
type PostModel struct {
	Ib     uint
	Thread uint
	ID     uint
	Result PostType
}

// PostType is the top level of the JSON response
type PostType struct {
	Body Post `json:"post"`
}

// Post holds the post information
type Post struct {
	ThreadID    uint       `json:"thread_id"`
	PostID      uint       `json:"post_id"`
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
func (i *PostModel) Get() (err error) {

	if i.Ib == 0 || i.Thread == 0 || i.ID == 0 {
		return e.ErrNotFound
	}

	// Initialize response header
	response := PostType{}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	post := Post{}

	// SQL query to fetch post details along with associated user and image information.
	// The query joins the posts table with images, threads, and users tables.
	// It also includes a subquery to get the maximum role_id for the user in the specific ib_id.
	// The query filters out deleted threads and posts.
	err = dbase.QueryRow(`
        SELECT 
            threads.thread_id,
            posts.post_id,
            post_num,
            user_name,
            users.user_id,
            COALESCE(
                (SELECT MAX(role_id) 
                 FROM user_ib_role_map 
                 WHERE user_ib_role_map.user_id = users.user_id 
                 AND ib_id = ?),
                user_role_map.role_id
            ) AS role,
            post_time,
            post_text,
            image_id,
            image_file,
            image_thumbnail,
            image_tn_height,
            image_tn_width
        FROM 
            posts
        LEFT JOIN 
            images ON posts.post_id = images.post_id
        INNER JOIN 
            threads ON posts.thread_id = threads.thread_id
        INNER JOIN 
            users ON posts.user_id = users.user_id
        INNER JOIN 
            user_role_map ON user_role_map.user_id = users.user_id
        WHERE 
            posts.post_num = ? 
            AND posts.thread_id = ? 
            AND threads.ib_id = ? 
            AND thread_deleted != 1 
            AND post_deleted != 1`,
		i.Ib, i.ID, i.Thread, i.Ib).Scan(
		&post.ThreadID,
		&post.PostID,
		&post.Num,
		&post.Name,
		&post.UID,
		&post.Group,
		&post.Time,
		&post.Text,
		&post.ImageID,
		&post.File,
		&post.Thumb,
		&post.ThumbHeight,
		&post.ThumbWidth,
	)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return
	}

	// Add pagedresponse to the response struct
	response.Body = post

	// This is the data we will serialize
	i.Result = response

	return

}
