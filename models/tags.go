package models

import (
	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"

	u "github.com/eirka/eirka-get/utils"
)

// TagsModel holds the parameters from the request and also the key for the cache
type TagsModel struct {
	Ib     uint
	Page   uint
	Result TagsType
}

// TagsType is the top level of the JSON response
type TagsType struct {
	Body u.PagedResponse `json:"tags"`
}

// Tags struct
type Tags struct {
	ID    uint   `json:"id"`
	Tag   string `json:"tag"`
	Total uint   `json:"total"`
	Type  uint   `json:"type"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *TagsModel) Get() (err error) {

	if i.Ib == 0 || i.Page == 0 {
		return e.ErrNotFound
	}

	// Initialize response header
	response := TagsType{}

	// tags slice
	tags := []Tags{}

	// Initialize struct for pagination
	paged := u.PagedResponse{}
	// Set current page to parameter
	paged.CurrentPage = i.Page
	// Set threads per index page to config setting
	paged.PerPage = config.Settings.Limits.PostsPerPage

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// This SQL query counts the total number of tags for the given image board (ib_id).
	// It's used to populate the pagination information.
	err = dbase.QueryRow("SELECT COUNT(*) FROM tags WHERE ib_id = ?", i.Ib).Scan(&paged.Total)
	if err != nil {
		return
	}

	// Calculate Limit and total Pages
	paged.Get()

	// Return 404 if page requested is larger than actual pages
	if i.Page > paged.Pages {
		return e.ErrNotFound
	}

	// This complex SQL query retrieves tag information along with image counts.
	// It joins multiple tables (tagmap, images, posts, threads) to ensure only
	// non-deleted content is counted. The results are grouped by tag_id,
	// ordered by count (descending) and tag_id (ascending), and limited for pagination.
	rows, err := dbase.Query(`
		SELECT IFNULL(tag_counts.count, 0) AS count, t.tag_id, t.tag_name, t.tagtype_id
		FROM tags t
		LEFT JOIN (
		SELECT tm.tag_id, COUNT(DISTINCT tm.image_id) as count
		FROM tagmap tm
		INNER JOIN images i ON tm.image_id = i.image_id
		INNER JOIN posts p ON i.post_id = p.post_id AND p.post_deleted != 1
		INNER JOIN threads th ON p.thread_id = th.thread_id AND th.thread_deleted != 1
		GROUP BY tm.tag_id
		) tag_counts ON t.tag_id = tag_counts.tag_id
		WHERE t.ib_id = ?
		ORDER BY count DESC, t.tag_id ASC
		LIMIT ?, ?`, i.Ib, paged.Limit, paged.PerPage)
	if err != nil {
		return
	}

	for rows.Next() {
		// Initialize posts struct
		tag := Tags{}
		// Scan rows and place column into struct
		err := rows.Scan(&tag.Total, &tag.ID, &tag.Tag, &tag.Type)
		if err != nil {
			rows.Close() // Explicitly close rows before returning
			return err
		}

		// Append rows to info struct
		tags = append(tags, tag)
	}
	if err = rows.Err(); err != nil {
		return err
	}

	// Add threads slice to items interface
	paged.Items = tags

	// Add pagedresponse to the response struct
	response.Body = paged

	// This is the data we will serialize
	i.Result = response

	return

}
