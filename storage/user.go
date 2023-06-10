package storage

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

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
	users   map[string]*User
	dataDir string
}

func NewUserStore(issuer string, dataDir string) (UserStore, error) {
	store := userStore{
		users:   make(map[string]*User),
		dataDir: dataDir,
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
	u.users = make(map[string]*User) // Clear the existing user dictionary

	files, err := os.ReadDir(u.dataDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			filePath := filepath.Join(u.dataDir, file.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				return err
			}

			var uu map[string]*User
			err = json.Unmarshal(data, &uu)
			if err != nil {
				return fmt.Errorf("invalid users in %s: %w", filePath, err)
			}

			for username, user := range uu {
				u.users[username] = user
			}

			log.Printf("loaded users from %s", filePath)
		}
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
