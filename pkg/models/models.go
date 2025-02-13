package models

import (
	"errors"
	"time"
)

var (
	ErrNoRecord = errors.New("models: no matching record found")
	// ErrInvalidCredentials Add a new ErrInvalidCredentials error. We'll use this later if a user
	// tries to log in with an incorrect email address or password.
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	// ErrDuplicateEmail Add a new ErrDuplicateEmail error. We'll use this later if a user
	// tries to signup with an email address that's already in use.
	ErrDuplicateEmail = errors.New("models: duplicate email")
)

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

// User Define a new User type. Notice how the field names and types align
// with the columns in the database `users` table?
type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}
