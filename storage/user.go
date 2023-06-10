package storage

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/text/language"
)

type User struct {
	ID                string       `json:"id"`
	Username          string       `json:"username"`
	Password          string       `json:"password"`
	FirstName         string       `json:"firstName"`
	LastName          string       `json:"lastName"`
	Email             string       `json:"email"`
	EmailVerified     bool         `json:"emailVerified"`
	Phone             string       `json:"phone"`
	PhoneVerified     bool         `json:"phoneVerified"`
	PreferredLanguage language.Tag `json:"preferredLanguage"`
	IsAdmin           bool         `json:"isAdmin"`
}

type Service struct {
	keys map[string]*rsa.PublicKey
}

type UserStore interface {
	GetUserByID(string) *User
	GetUserByUsername(string) *User
	ExampleClientID() string
}

type userStore struct {
	users    map[string]*User
	jsonFile io.Reader
}

func NewUserStore(issuer string, jsonFile io.Reader) (UserStore, error) {
	store := userStore{
		users:    make(map[string]*User),
		jsonFile: jsonFile,
	}

	err := store.LoadUsersFromJSON()
	if err != nil {
		return nil, fmt.Errorf("could not load users from JSON: %w", err)
	}

	return &store, nil
}

// ExampleClientID is only used in the example server
func (u *userStore) ExampleClientID() string {
	return "service"
}

func (u *userStore) LoadUsersFromJSON() error {
	data, err := io.ReadAll(u.jsonFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &u.users)
	if err != nil {
		return err
	}

	return nil
}

func (u *userStore) GetUserByID(id string) *User {
	for _, user := range u.users {
		if user.ID == id {
			return user
		}
	}
	return nil
}

func (u *userStore) GetUserByUsername(username string) *User {
	return u.users[username]
}
