package models

import (
	"database/sql"

	e "github.com/techjanitor/pram-get/errors"
	u "github.com/techjanitor/pram-get/utils"
)

// ImageModel holds the parameters from the request and also the key for the cache
type ImageModel struct {
	User   uint
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
	Prev    uint        `json:"prev,omitempty"`
	Next    uint        `json:"next,omitempty"`
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

	// Initialize response header
	response := ImageType{}

	// Get Database handle
	db, err := u.GetDb()
	if err != nil {
		return
	}

	imageheader := ImageHeader{}

	err = db.QueryRow(`SELECT image_id,posts.thread_id,posts.post_num,posts.post_id,image_file,image_orig_height,image_orig_width FROM images
        LEFT JOIN posts on images.post_id = posts.post_id
        LEFT JOIN threads on posts.thread_id = threads.thread_id
        WHERE image_id = ? AND ib_id = ?`, i.Id, i.Ib).Scan(&imageheader.Id, &imageheader.Thread, &imageheader.PostNum, &imageheader.PostId, &imageheader.File, &imageheader.Height, &imageheader.Width)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return
	}

	// Check to see if image has been deleted
	if u.GetBool("image_deleted", "images", "image_id", i.Id) {
		return e.ErrNotFound
	}

	var image_num, image_total uint

	// Get total amount of images in the thread
	err = db.QueryRow(`SELECT count(image_id)
	FROM images 
	LEFT JOIN posts on images.post_id = posts.post_id 
	LEFT JOIN threads on posts.thread_id = threads.thread_id 
	WHERE threads.thread_id = ?`, imageheader.Thread).Scan(&image_total)
	if err != nil {
		return
	}

	// Get the position of the image in the thread
	err = db.QueryRow(`SELECT image_num FROM
	(SELECT image_id, @rownum := @rownum + 1 AS image_num
	FROM images 
	JOIN (SELECT @rownum := 0) r
	LEFT JOIN posts on images.post_id = posts.post_id 
	LEFT JOIN threads on posts.thread_id = threads.thread_id 
	WHERE threads.thread_id = ?
	ORDER BY images.post_id) as F
	WHERE image_id = ?`, imageheader.Thread, i.Id).Scan(&image_num)
	if err != nil {
		return
	}

	if image_num > 1 {
		prev_image := image_num - 1

		// Get the previous image id
		err = db.QueryRow(`SELECT image_id FROM
		(SELECT image_id, @rownum := @rownum + 1 AS rank
		FROM images 
		JOIN (SELECT @rownum := 0) r
		LEFT JOIN posts on images.post_id = posts.post_id 
		LEFT JOIN threads on posts.thread_id = threads.thread_id 
		WHERE threads.thread_id = ?
		ORDER BY images.post_id) as F
		WHERE rank = ?`, imageheader.Thread, prev_image).Scan(&imageheader.Prev)
		if err != nil {
			return
		}

	}

	if image_num < image_total {
		next_image := image_num + 1

		// Get the next image id
		err = db.QueryRow(`SELECT image_id FROM
		(SELECT image_id, @rownum := @rownum + 1 AS rank
		FROM images 
		JOIN (SELECT @rownum := 0) r
		LEFT JOIN posts on images.post_id = posts.post_id 
		LEFT JOIN threads on posts.thread_id = threads.thread_id 
		WHERE threads.thread_id = ?
		ORDER BY images.post_id) as F
		WHERE rank = ?`, imageheader.Thread, next_image).Scan(&imageheader.Next)
		if err != nil {
			return
		}

	}

	// Get tags for image
	rows, err := db.Query("SELECT tags.tag_id,tagtype_id,tag_name from tagmap LEFT JOIN tags on tagmap.tag_id = tags.tag_id WHERE image_id = ?", i.Id)
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
