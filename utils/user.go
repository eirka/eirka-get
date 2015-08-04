package utils

import (
	"errors"
	"fmt"

	e "github.com/techjanitor/pram-get/errors"
)

var (
	ErrUserNotConfirmed error = errors.New("Account not confirmed")
	ErrUserBanned       error = errors.New("Account banned")
	ErrUserLocked       error = errors.New("Account locked")
	userdataWorker      *userWorker
)

// struct for database insert worker
type userWorker struct {
	send    chan *User
	receive chan *User
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
}

func init() {
	// make worker channel
	userdataWorker = &userWorker{
		make(chan *User, 64),
		make(chan *User, 64),
	}

	go func() {

		// Get Database handle
		db, err := GetDb()
		if err != nil {
			return
		}

		// prepare query for users table
		ps1, err := db.Prepare("SELECT usergroup_id,user_name,user_email,user_confirmed,user_locked,user_banned FROM users WHERE user_id = ?")
		if err != nil {
			return
		}

		// range through tasks channel
		for u := range userdataWorker.send {

			// input data
			err = ps1.QueryRow(u.Id).Scan(&u.Group, &u.Name, &u.Email, &u.IsConfirmed, &u.IsLocked, &u.IsBanned)
			if err != nil {
				return
			}

			userdataWorker.receive <- u

		}

	}()

}

// get the user info from id
func (u *User) Info() (err error) {

	// this needs an id
	if u.Id == 0 || u.Id == 1 {
		return e.ErrInvalidParam
	}

	userdataWorker.send <- u

	fmt.Printf("%s\n", u)

	<-userdataWorker.receive

	fmt.Printf("%s\n", u)

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
