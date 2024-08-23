package models

import (
	"fmt"
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

// remove bad characters
func formatQuery(str string) string {
	return strings.Map(func(r rune) rune {
		if !strings.ContainsRune(`"'+-@><()~*`, r) {
			return r
		}
		return -1
	}, str)
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
		// get rid of bad cahracters for mysql in boolean mode
		term = formatQuery(term)

		exact = append(exact, term)
		searchquery = append(searchquery, fmt.Sprintf("+%s", term))
	}

	// This SQL query performs a complex search for tags based on the given terms.
	// It does the following:
	// 1. Counts the number of images associated with each tag, considering only non-deleted posts and threads.
	// 2. Retrieves tag information (id, name, type).
	// 3. Calculates two scores:
	//    - An exact match score (1 if the tag name exactly matches the search terms, 0 otherwise)
	//    - A relevance score using MySQL's full-text search capabilities
	// 4. Filters tags based on the search terms using MySQL's boolean mode full-text search
	// 5. Orders the results by exact match score (descending) and then by relevance score (descending)
	rows, err := dbase.Query(`
		SELECT count, tag_id, tag_name, tagtype_id
		FROM (
			SELECT 
				(SELECT COUNT(tagmap.image_id) 
				 FROM tagmap
				 INNER JOIN images ON tagmap.image_id = images.image_id
				 INNER JOIN posts ON images.post_id = posts.post_id
				 INNER JOIN threads ON posts.thread_id = threads.thread_id
				 WHERE tagmap.tag_id = tags.tag_id 
				   AND post_deleted != 1 
				   AND thread_deleted != 1
				) AS count,
				tag_id, tag_name, tagtype_id,
				CASE WHEN tag_name = ? THEN 1 ELSE 0 END AS score,
				MATCH(tag_name) AGAINST (? IN BOOLEAN MODE) AS score2
			FROM tags 
			WHERE MATCH(tag_name) AGAINST (? IN BOOLEAN MODE) 
			  AND ib_id = ?
			GROUP BY tag_id 
			ORDER BY score DESC, score2 DESC
		) AS search`,
		strings.Join(exact, " "),
		strings.Join(searchquery, " "),
		fmt.Sprintf("%s*", strings.Join(searchquery, " ")),
		i.Ib,
	)
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

	// Add pagedresponse to the response struct
	response.Body = tags

	// This is the data we will serialize
	i.Result = response

	return

}
