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

	// Process valid terms (after cleaning)
	var validTerms []string
	for _, term := range terms {
		// Clean term for MySQL boolean mode
		term = u.FormatQuery(term)
		if term == "" {
			continue // Skip empty terms
		}
		validTerms = append(validTerms, term)
	}

	// Build query parameters in correct order for the SQL query
	var params []interface{}
	
	// 1. Prepare the CASE expression parameter (exact match comparison)
	fullSearchTerm := strings.TrimSpace(i.Term)
	params = append(params, fullSearchTerm)

	// Define the CASE statement for exact matches
	exactMatchCase := "CASE WHEN tag_name = ? THEN 1 ELSE 0 END"

	// Construct the search expressions
	var booleanMatchExpr, booleanWhereExpr string
	
	if len(validTerms) == 0 {
		// No valid search terms
		booleanMatchExpr = "MATCH(tag_name) AGAINST ('' IN BOOLEAN MODE)"
		booleanWhereExpr = "MATCH(tag_name) AGAINST ('' IN BOOLEAN MODE)"
	} else {
		// We have valid terms - add them directly to the parameter list
		// Format for relevance scoring (used in score2)
		relevanceSearch := "+" + strings.Join(validTerms, " +") 
		params = append(params, relevanceSearch)
		booleanMatchExpr = "MATCH(tag_name) AGAINST (? IN BOOLEAN MODE)"
		
		// Format for wildcard searching (used in WHERE clause)
		wildcardSearch := "+" + strings.Join(validTerms, "* +") + "*"
		params = append(params, wildcardSearch)
		booleanWhereExpr = "MATCH(tag_name) AGAINST (? IN BOOLEAN MODE)"
	}
	
	// Add the image board parameter as the last parameter
	params = append(params, i.Ib)

	// This SQL query performs a complex search for tags based on the given terms.
	// The parameter order matters and must match the sequence in the params slice:
	// 1. exactMatch parameter for the CASE expression
	// 2. relevanceSearch parameter for booleanMatchExpr (score2)
	// 3. wildcardSearch parameter for booleanWhereExpr (WHERE clause)
	// 4. image board ID for the final filter
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