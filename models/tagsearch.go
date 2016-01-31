package models

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/eirka/eirka-libs/validate"
)

// TagSearchModel holds the parameters from the request and also the key for the cache
type TagSearchModel struct {
	Ib     uint
	Term   string
	Result TagSearchType
}

// TagSearchType is the top level of the JSON response
type TagSearchType struct {
	Body []Tags `json:"tagsearch"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *TagSearchModel) Get() (err error) {

	// Initialize response header
	response := TagSearchType{}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	tags := []Tags{}

	tag := validate.Validate{Input: i.Term, Max: config.Settings.Limits.TagMaxLength, Min: config.Settings.Limits.TagMinLength}
	if tag.IsEmpty() {
		return e.ErrNoTagName
	} else if tag.MinLength() {
		return e.ErrTagShort
	} else if tag.MaxLength() {
		return e.ErrTagLong
	}

	terms := strings.Split(strings.TrimSpace(i.Term), " ")

	var exact, searchquery []string

	for _, term := range terms {

		// wrap in quotes
		term = strconv.Quote(term)

		exact = append(exact, term)
		searchquery = append(searchquery, fmt.Sprintf("+%s", term))
	}

	rows, err := dbase.Query(`SELECT count,tag_id,tag_name,tagtype_id
    FROM (SELECT (SELECT count(tagmap.image_id) FROM tagmap
    INNER JOIN images on tagmap.image_id = images.image_id
    INNER JOIN posts on images.post_id = posts.post_id 
    INNER JOIN threads on posts.thread_id = threads.thread_id 
    WHERE tagmap.tag_id = tags.tag_id AND post_deleted != 1 AND thread_deleted != 1) as count,
    tag_id,tag_name,tagtype_id,
    CASE WHEN tag_name = ? THEN 1 ELSE 0 END AS score, 
    MATCH(tag_name) AGAINST (? IN BOOLEAN MODE) AS score2
    FROM tags WHERE MATCH(tag_name) AGAINST (? IN BOOLEAN MODE) AND ib_id = ?
    GROUP BY tag_id ORDER BY score DESC, score2 DESC) as search`, strings.Join(exact, " "), strings.Join(searchquery, " "), fmt.Sprintf("%s*", strings.Join(searchquery, " ")), i.Ib)
	if err != nil {
		return
	}

	for rows.Next() {
		// Initialize posts struct
		tag := Tags{}
		// Scan rows and place column into struct
		err := rows.Scan(&tag.Total, &tag.Id, &tag.Tag, &tag.Type)
		if err != nil {
			return err
		}

		// Append rows to info struct
		tags = append(tags, tag)
	}
	err = rows.Err()
	if err != nil {
		return
	}

	// Add pagedresponse to the response struct
	response.Body = tags

	// This is the data we will serialize
	i.Result = response

	return

}
