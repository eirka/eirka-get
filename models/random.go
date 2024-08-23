package models

import (
	"database/sql"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

// RandomModel holds the parameters from the request and also the key for the cache
type RandomModel struct {
	Ib     uint
	Result ImageType
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *RandomModel) Get() (err error) {

	if i.Ib == 0 {
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
    WHERE ib_id = ? AND thread_deleted != 1 AND post_deleted != 1
    ORDER BY RAND() LIMIT 1`, i.Ib).Scan(&imageheader.ID, &imageheader.Thread, &imageheader.PostNum, &imageheader.PostID, &imageheader.File, &imageheader.Height, &imageheader.Width)
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
    ORDER BY images.post_id ASC LIMIT 1) as next`, imageheader.Thread, imageheader.ID, imageheader.Thread, imageheader.ID).Scan(&imageheader.Prev, &imageheader.Next)
	if err != nil {
		return
	}

	// Get tags for image
	rows, err := dbase.Query("SELECT tags.tag_id,tagtype_id,tag_name from tagmap LEFT JOIN tags on tagmap.tag_id = tags.tag_id WHERE image_id = ?", imageheader.ID)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		// Initialize posts struct
		tag := ImageTags{}
		// Scan rows and place column into struct
		err := rows.Scan(&tag.ID, &tag.Type, &tag.Tag)
		if err != nil {
			return err
		}
		// Append rows to info struct
		imageheader.Tags = append(imageheader.Tags, tag)
	}
	if err = rows.Err(); err != nil {
		return err
	}

	// Add pagedresponse to the response struct
	response.Body = imageheader

	// This is the data we will serialize
	i.Result = response

	return

}
