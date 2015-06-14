package models

import (
	"database/sql"

	e "github.com/techjanitor/pram-get/errors"
	u "github.com/techjanitor/pram-get/utils"
)

// TagInfoModel holds the parameters from the request and also the key for the cache
type TagInfoModel struct {
	Id     uint
	Result TagInfoType
}

// IndexType is the top level of the JSON response
type TagInfoType struct {
	Tag TagInfo `json:"taginfo"`
}

// Header for tag page
type TagInfo struct {
	Id   uint   `json:"id"`
	Tag  string `json:"tag"`
	Type uint   `json:"type"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *TagInfoModel) Get() (err error) {

	// Initialize response header
	response := TagInfoType{}

	// Initialize struct for tag info
	taginfo := TagInfo{}

	taginfo.Id = i.Id

	// Get Database handle
	db, err := u.GetDb()
	if err != nil {
		return
	}

	// Get tag name and type
	err = db.QueryRow("select tag_name,tagtype_id from tags where tag_id = ?", i.Id).Scan(&taginfo.Tag, &taginfo.Type)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return
	}

	// Add taginfo to the response struct
	response.Tag = taginfo

	// This is the data we will serialize
	i.Result = response

	return

}
