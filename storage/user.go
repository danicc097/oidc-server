package storage

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
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
					log.Println("%s: %s: user with ID %s already exists", filePath, user.Username, user.ID)
				}
				u.users[user.ID] = user
			}

			log.Println("loaded users from %s", filePath)
		}
	}

	return nil
}

func (u userStore) GetUserByID(id string) *User {
	u.mu.RLock()
	defer u.mu.RUnlock()

	return u.users[id]
}

func (u userStore) GetUserByUsername(username string) *User {
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
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("file modified:", event.Name)
					err := userStore.LoadUsersFromJSON()
					if err != nil {
						// don't shutdown server. if we messed up just check the logs...
						// there should be a global StoreErrors []string with rw mutex
						// so that if len > 0 we render an error message in login page, callbacks...
						// this way we notice the error asap without checking logs.
						log.Println("error reloading users: %s", err)
					}
					// if ok, set StoreErrors to empty slice
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("watcher error:", err)
			}
		}
	}()

	err = filepath.Walk(dataDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Println("walkDir error:", err)
			return err
		}
		if !info.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				log.Println("watcher error:", err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Println("walk error:", err)
	}

	<-done
}
