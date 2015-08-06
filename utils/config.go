package utils

import (
	"github.com/techjanitor/pram-get/config"
)

// Get limits that are in the database
func GetDatabaseSettings() {

	// Get Database handle
	db, err := GetDb()
	if err != nil {
		panic(err)
	}

	ps, err := db.Prepare("SELECT settings_value FROM settings WHERE settings_key = ? LIMIT 1")
	if err != nil {
		panic(err)
	}
	defer ps.Close()

	err = ps.QueryRow("antispam_cookiename").Scan(&config.Settings.Antispam.CookieName)
	if err != nil {
		panic(err)
	}

	err = ps.QueryRow("antispam_cookievalue").Scan(&config.Settings.Antispam.CookieValue)
	if err != nil {
		panic(err)
	}

	err = ps.QueryRow("thread_postsperpage").Scan(&config.Settings.Limits.PostsPerPage)
	if err != nil {
		panic(err)
	}

	err = ps.QueryRow("index_threadsperpage").Scan(&config.Settings.Limits.ThreadsPerPage)
	if err != nil {
		panic(err)
	}

	err = ps.QueryRow("index_postsperthread").Scan(&config.Settings.Limits.PostsPerThread)
	if err != nil {
		panic(err)
	}

	err = ps.QueryRow("tag_maxlength").Scan(&config.Settings.Limits.TagMaxLength)
	if err != nil {
		panic(err)
	}

	err = ps.QueryRow("tag_minlength").Scan(&config.Settings.Limits.TagMinLength)
	if err != nil {
		panic(err)
	}

	err = ps.QueryRow("param_maxsize").Scan(&config.Settings.Limits.ParamMaxSize)
	if err != nil {
		panic(err)
	}

	err = ps.QueryRow("hmac_secret").Scan(&config.Settings.Session.Secret)
	if err != nil {
		panic(err)
	}

	return

}
