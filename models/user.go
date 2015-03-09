package models

import (
	"encoding/json"
)

// User represents an user within the authentication system
type User struct {
	// ID represents the database ID of the User
	ID int64 `json:"id"`
	// Username represents the username of the User
	Username string `json:"username"`
	// Password represents the bcrypt-hashed password of the User
	Password string `json:"-"`
	// Email represents the email address of the User
	Email string `json:"email"`
	// VerifiedEmail indicates whether the user has verified their email address
	VerifiedEmail bool `json:"verifiedEmail"`
	// Active indicates whether the User is set as active
	Active bool `json:"active"`
}

// NewUser creates a new user with the given information
func NewUser(username string, password string, email string, verified bool, active bool) *User {
	user := &User{
		ID:            -1,
		Username:      username,
		Password:      password,
		Email:         email,
		VerifiedEmail: verified,
		Active:        active,
	}

	return user
}

// String represents a JSON encoded representation of the user
func (user *User) String() string {
	jsonContent, err := json.Marshal(user)
	if err != nil {
		return ""
	}

	return string(jsonContent)
}
