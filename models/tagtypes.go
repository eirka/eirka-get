package models

import (
	"github.com/eirka/eirka-libs/db"
)

// TagTypesModel holds the parameters from the request and also the key for the cache
type TagTypesModel struct {
	Result TagTypesType
}

// TagTypesType is the top level of the JSON response
type TagTypesType struct {
	Body []TagTypes `json:"tagtypes"`
}

// TagTypes truct
type TagTypes struct {
	ID   uint   `json:"id"`
	Type string `json:"type"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *TagTypesModel) Get() (err error) {

	// Initialize response header
	response := TagTypesType{}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	tags := []TagTypes{}

	rows, err := dbase.Query("select tagtype_id,tagtype_name from tagtype")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		// Initialize posts struct
		tag := TagTypes{}
		// Scan rows and place column into struct
		err := rows.Scan(&tag.ID, &tag.Type)
		if err != nil {
			return err
		}
		// Append rows to info struct
		tags = append(tags, tag)
	}
	if rows.Err() != nil {
		return
	}

	// Add pagedresponse to the response struct
	response.Body = tags

	// This is the data we will serialize
	i.Result = response

	return

}
