package utils

import (
	"database/sql"
	"errors"
	"fmt"

	e "github.com/techjanitor/pram-get/errors"
)

var (
	ErrUserNotConfirmed error = errors.New("Account not confirmed")
	ErrUserBanned       error = errors.New("Account banned")
	ErrUserLocked       error = errors.New("Account locked")
	ErrUserNotExist     error = errors.New("User does not exist")
	userdataWorker      *userWorker
)

// struct for database insert worker
type userWorker struct {
	send chan<- *User
	done <-chan bool
}

// user struct
type User struct {
	Id              uint   `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Group           uint   `json:"group"`
	IsConfirmed     bool   `json:"-"`
	IsLocked        bool   `json:"-"`
	IsBanned        bool   `json:"-"`
	IsAuthenticated bool   `json:"-"`
	err             error  `json:"-"`
}

func UserInit() {

	fmt.Println("Initializing User Info Worker...")

	// make worker channel
	userdataWorker = &userWorker{
		send: make(chan *User),
		done: make(chan bool),
	}

	go func() {

		// Get Database handle
		db, err := GetDb()
		if err != nil {
			fmt.Println(err)
			return
		}

		// prepare query for users table
		ps1, err := db.Prepare("SELECT usergroup_id,user_name,user_email,user_confirmed,user_locked,user_banned FROM users WHERE user_id = ?")
		if err != nil {
			fmt.Println(err)
			return
		}

		// range through tasks channel
		for u := range userdataWorker.send {

			// input data
			err = ps1.QueryRow(u.Id).Scan(&u.Group, &u.Name, &u.Email, &u.IsConfirmed, &u.IsLocked, &u.IsBanned)
			if err != nil {
				u.err = err
			}

			userdataWorker.done <- true

		}

	}()

}

// get the user info from id
func (u *User) Info() (err error) {

	// this needs an id
	if u.Id == 0 || u.Id == 1 {
		return e.ErrInvalidParam
	}

	// get original uid
	uid := u.Id

	fmt.Printf("first: %s\n", u)

	// send to worker
	userdataWorker.send <- u

	// block until done
	<-userdataWorker.done

	fmt.Printf("second: %s\n", u)

	// check error
	if u.err == sql.ErrNoRows {
		return ErrUserNotExist
	} else if u.err != nil {
		return e.ErrInternalError
	}

	// check that theyre still the same just in case
	if u.Id != uid {
		return e.ErrInternalError
	}

	// if account is not confirmed
	if !u.IsConfirmed {
		return ErrUserNotConfirmed
	}

	// if locked
	if u.IsLocked {
		return ErrUserLocked
	}

	// if banned
	if u.IsBanned {
		return ErrUserBanned
	}

	// mark authenticated
	u.IsAuthenticated = true

	return

}
