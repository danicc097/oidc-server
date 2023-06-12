package storage

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
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

// TODO: User struct should be defined by client using this server
// and use generics to fill in.
// func (s *Storage) setUserinfo should instead be passed as arg to NewStorage.
// signature can remain func(ctx context.Context, userInfo *oidc.UserInfo, userID, clientID string, scopes []string) (err error)
// and if not set, use default setUserInfo.
// due to this, we need to expose main.go as a Run function that accepts the custom store
// so that clients just use oidcserver.Run(store)
// and leave dockerfile up to clients
type UserStore interface {
	GetUserByID(string) *User
	GetUserByUsername(string) *User
	ExampleClientID() string
}

type userStore struct {
	users   map[string]*User
	dataDir string
	mu      sync.RWMutex
}

var StorageErrors struct {
	Errors []string
	mu     sync.RWMutex
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

	go watchUsersFolder(dataDir, &store)

	return &store, nil
}

func (u *userStore) ExampleClientID() string {
	return "service"
}

func (u *userStore) LoadUsersFromJSON() error {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.users = make(map[string]*User)

	files, err := os.ReadDir(u.dataDir)
	if err != nil {
		return err
	}

	errs := []string{}

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

			for _, user := range uu {
				if _, exists := u.users[user.ID]; exists {
					errMsg := fmt.Sprintf("%s: %s: user with ID %s already exists", filePath, user.Username, user.ID)
					errs = append(errs, errMsg)
					log.Println(errMsg)
				}
				u.users[user.ID] = user
			}

			if len(errs) > 0 {
				return errors.New(strings.Join(errs, "\n"))
			}

			log.Printf("loaded users from %s", filePath)
		}
	}

	return nil
}

func (u *userStore) GetUserByID(id string) *User {
	u.mu.RLock()
	defer u.mu.RUnlock()

	return u.users[id]
}

func (u *userStore) GetUserByUsername(username string) *User {
	u.mu.RLock()
	defer u.mu.RUnlock()

	for _, user := range u.users {
		if user.Username == username {
			return user
		}
	}
	return nil
}

func watchUsersFolder(dataDir string, userStore *userStore) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					log.Printf("file modified: %s", event.Name)
					err := userStore.LoadUsersFromJSON()
					StorageErrors.mu.Lock()
					StorageErrors.Errors = []string{}
					if err != nil {
						errMsg := fmt.Sprintf("error reloading users: %s", err)
						StorageErrors.Errors = append(StorageErrors.Errors, errMsg)
						log.Println(errMsg)
					}
					StorageErrors.mu.Unlock()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("watcher error: %s", err)
			}
		}
	}()

	err = filepath.WalkDir(dataDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("walkDir error: %s", err)
			return err
		}
		if !d.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				log.Printf("watcher error: %s", err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("walk error: %s", err)
	}

	<-done
}
