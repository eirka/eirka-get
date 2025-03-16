package models

import (
	"strings"

	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/eirka/eirka-libs/validate"

	u "github.com/eirka/eirka-get/utils"
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

// This function has been moved to utils.FormatQuery

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

	// Build query and parameters safely with each term as a separate parameter
	var params []interface{}
	var placeholders, booleanPlaceholders []string

	// Prepare basic parts of the query
	exactMatchCase := "CASE WHEN tag_name = ? THEN 1 ELSE 0 END"

	// Construct dynamic parts based on number of search terms
	fullSearchTerm := strings.TrimSpace(i.Term)

	// Add parameters for the queries
	// First parameter: exact match comparison
	params = append(params, fullSearchTerm)

	// Build the boolean mode search string with placeholders
	for _, term := range terms {
		// Clean term for MySQL boolean mode
		term = u.FormatQuery(term)
		if term == "" {
			continue // Skip empty terms
		}

		// For boolean search relevance scoring
		placeholders = append(placeholders, "+?")
		params = append(params, term)

		// For boolean search with wildcard
		booleanPlaceholders = append(booleanPlaceholders, "+?*")
		params = append(params, term)
	}

	// Add the image board parameter
	params = append(params, i.Ib)

	// Construct the search expressions
	booleanMatchExpr := "MATCH(tag_name) AGAINST (CONCAT(" + strings.Join(placeholders, ", ' ', ") + ") IN BOOLEAN MODE)"
	booleanWhereExpr := "MATCH(tag_name) AGAINST (CONCAT(" + strings.Join(booleanPlaceholders, ", ' ', ") + ") IN BOOLEAN MODE)"

	// This SQL query performs a complex search for tags based on the given terms.
	// Now using proper parameterization for security:
	// 1. Counts the number of images associated with each tag, considering only non-deleted posts and threads.
	// 2. Retrieves tag information (id, name, type).
	// 3. Calculates two scores:
	//    - An exact match score (1 if the tag name exactly matches the search terms, 0 otherwise)
	//    - A relevance score using MySQL's full-text search with parameterized inputs
	// 4. Filters tags based on the search terms using MySQL's boolean mode full-text search with parameterized inputs
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
				`+exactMatchCase+` AS score,
				`+booleanMatchExpr+` AS score2
			FROM tags 
			WHERE `+booleanWhereExpr+` 
			  AND ib_id = ?
			GROUP BY tag_id 
			ORDER BY score DESC, score2 DESC
		) AS search`,
		params...,
	)
	if err != nil {
		return
	}
	// Ensure rows are closed even if there's an error later in the function
	defer rows.Close()

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
