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
	Id     uint
	Result PostType
}

// IndexType is the top level of the JSON response
type PostType struct {
	Body Post `json:"post"`
}

type Post struct {
	ThreadId    uint       `json:"thread_id"`
	PostId      uint       `json:"post_id"`
	Num         uint       `json:"num"`
	Name        string     `json:"name"`
	Uid         uint       `json:"uid"`
	Group       uint       `json:"group"`
	Time        *time.Time `json:"time"`
	Text        *string    `json:"comment"`
	ImgId       *uint      `json:"img_id,omitempty"`
	File        *string    `json:"filename,omitempty"`
	Thumb       *string    `json:"thumbnail,omitempty"`
	ThumbHeight *uint      `json:"tn_height,omitempty"`
	ThumbWidth  *uint      `json:"tn_width,omitempty"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *PostModel) Get() (err error) {

	if i.Ib == 0 || i.Thread == 0 || i.Id == 0 {
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

	err = dbase.QueryRow(`SELECT threads.thread_id,posts.post_id,post_num,user_name,user_id
    COALESCE((SELECT MAX(role_id) FROM user_ib_role_map WHERE user_ib_role_map.user_id = users.user_id AND ib_id = ?),user_role_map.role_id) as role,
    post_time,post_text,image_id,image_file,image_thumbnail,image_tn_height,image_tn_width
    FROM posts
    LEFT JOIN images on posts.post_id = images.post_id
    INNER JOIN threads on posts.thread_id = threads.thread_id
    INNER JOIN users on posts.user_id = users.user_id
    INNER JOIN user_role_map ON (user_role_map.user_id = users.user_id)
    WHERE posts.post_num = ? AND posts.thread_id = ? AND threads.ib_id = ? AND thread_deleted != 1 AND post_deleted != 1`, i.Ib, i.Id, i.Thread, i.Ib).Scan(&post.ThreadId, &post.PostId, &post.Num, &post.Name, &post.Uid, &post.Group, &post.Time, &post.Text, &post.ImgId, &post.File, &post.Thumb, &post.ThumbHeight, &post.ThumbWidth)
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
