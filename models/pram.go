package models

import (
	"github.com/techjanitor/pram-get/config"
	e "github.com/techjanitor/pram-get/errors"
	u "github.com/techjanitor/pram-get/utils"
)

// PramModel holds the parameters from the request and also the key for the cache
type PramModel struct {
	Result PramType
}

// ImageboardsType is the top level of the JSON response
type PramType struct {
	Info Pram         `json:"pram"`
	Body []Imageboard `json:"imageboards"`
}

// Info has information about pram
type Pram struct {
	Version string `json:"version"`
}

// Imageboard has information and statistics about the boards on a pram
type Imageboard struct {
	Id          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Domain      string `json:"url"`
	Threads     uint   `json:"threads"`
	Posts       uint   `json:"posts"`
	Images      uint   `json:"images"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *PramModel) Get() (err error) {

	// Initialize response header
	response := PramType{}

	// Add pram info to response
	pram := Pram{
		Version: config.PramVersion,
	}

	response.Info = pram

	// Get Database handle
	db, err := u.GetDb()
	if err != nil {
		return
	}

	rows, err := db.Query(`SELECT ib_id, ib_title, ib_description, ib_domain, 
	(SELECT COUNT(thread_id)
	FROM threads
	WHERE threads.ib_id=imageboards.ib_id) AS thread_count,
	(SELECT COUNT(post_id)
	FROM threads
	LEFT JOIN posts ON posts.thread_id = threads.thread_id
	WHERE threads.ib_id=imageboards.ib_id) AS post_count,
	(SELECT COUNT(image_id)
	FROM threads
	LEFT JOIN posts ON posts.thread_id = threads.thread_id
	LEFT JOIN images ON images.post_id = posts.post_id
	WHERE threads.ib_id=imageboards.ib_id) AS image_count
	FROM imageboards`)
	if err != nil {
		return
	}
	defer rows.Close()

	boards := []Imageboard{}
	for rows.Next() {
		board := Imageboard{}
		err := rows.Scan(&board.Id, &board.Title, &board.Description, &board.Domain, &board.Threads, &board.Posts, &board.Images)
		if err != nil {
			return err
		}
		boards = append(boards, board)
	}
	err = rows.Err()
	if err != nil {
		return
	}

	// Return 404 if there are no threads in ib
	if len(boards) == 0 {
		return e.ErrNotFound
	}

	// Add pagedresponse to the response struct
	response.Body = boards

	// This is the data we will serialize
	i.Result = response

	return

}
