package models

import (
	"database/sql"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

// ImageModel holds the parameters from the request and also the key for the cache
type ImageModel struct {
	Ib     uint
	Id     uint
	Result ImageType
}

// IndexType is the top level of the JSON response
type ImageType struct {
	Body ImageHeader `json:"image"`
}

// Header for image page
type ImageHeader struct {
	Id      uint        `json:"id"`
	Thread  uint        `json:"thread"`
	PostNum uint        `json:"post_num"`
	PostId  uint        `json:"post_id"`
	Prev    *uint       `json:"prev,omitempty"`
	Next    *uint       `json:"next,omitempty"`
	Width   uint        `json:"width"`
	Height  uint        `json:"height"`
	File    string      `json:"filename"`
	Tags    []ImageTags `json:"tags,omitempty"`
}

// Tags for image page
type ImageTags struct {
	Id   uint   `json:"id"`
	Tag  string `json:"tag"`
	Type string `json:"type"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *ImageModel) Get() (err error) {

	if i.Ib == 0 || i.Id == 0 {
		return e.ErrNotFound
	}

	// Initialize response header
	response := ImageType{}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	imageheader := ImageHeader{}

	// get image information
	err = dbase.QueryRow(`SELECT image_id,posts.thread_id,posts.post_num,posts.post_id,image_file,image_orig_height,image_orig_width 
	FROM images
    INNER JOIN posts on images.post_id = posts.post_id
    INNER JOIN threads on posts.thread_id = threads.thread_id
    WHERE image_id = ? AND ib_id = ? AND thread_deleted != 1 AND post_deleted != 1`, i.Id, i.Ib).Scan(&imageheader.Id, &imageheader.Thread, &imageheader.PostNum, &imageheader.PostId, &imageheader.File, &imageheader.Height, &imageheader.Width)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return
	}

	// Get the next and previous image id
	err = dbase.QueryRow(`SELECT (SELECT image_id 
    FROM images 
    INNER JOIN posts on images.post_id = posts.post_id 
    INNER JOIN threads on posts.thread_id = threads.thread_id 
    WHERE threads.thread_id = ? AND post_deleted != 1 AND image_id < ?
    ORDER BY images.post_id DESC LIMIT 1) as previous,
    (SELECT image_id 
    FROM images 
    INNER JOIN posts on images.post_id = posts.post_id 
    INNER JOIN threads on posts.thread_id = threads.thread_id 
    WHERE threads.thread_id = ? AND post_deleted != 1 AND image_id > ?
    ORDER BY images.post_id ASC LIMIT 1) as next`, imageheader.Thread, i.Id, imageheader.Thread, i.Id).Scan(&imageheader.Prev, &imageheader.Next)
	if err != nil {
		return
	}

	// Get tags for image
	rows, err := dbase.Query("SELECT tags.tag_id,tagtype_id,tag_name from tagmap LEFT JOIN tags on tagmap.tag_id = tags.tag_id WHERE image_id = ?", i.Id)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		// Initialize posts struct
		tag := ImageTags{}
		// Scan rows and place column into struct
		err := rows.Scan(&tag.Id, &tag.Type, &tag.Tag)
		if err != nil {
			return err
		}
		// Append rows to info struct
		imageheader.Tags = append(imageheader.Tags, tag)
	}
	err = rows.Err()
	if err != nil {
		return
	}

	// Add pagedresponse to the response struct
	response.Body = imageheader

	// This is the data we will serialize
	i.Result = response

	return

}
