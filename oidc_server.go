/**
 * Package oidc_server is a modified version of the example server at https://github.com/zitadel/oidc/tree/main/example/server.
 */
package oidc_server

import (
	"context"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/danicc097/oidc-server/exampleop"
	"github.com/danicc097/oidc-server/storage"
)

// Config defines OIDC server configuration.
type Config struct {
	// SetUserInfoFunc overrides population of userinfo .
	SetUserInfoFunc storage.SetUserInfoFunc
	// TLS runs the server with the given certificate.
	TLS *struct {
		CertFile string
		KeyFile  string
	}
}

// Runs starts the OIDC server.
func Run(config Config) {
	ctx := context.Background()

	issuer := os.Getenv("OIDC_ISSUER")
	port := "10001" // for internal network
	usersDataDir := path.Join(os.Getenv("DATA_DIR"), "users")

	us, err := storage.NewUserStore(issuer, usersDataDir)
	if err != nil {
		log.Fatal("could not create user store: ", err)
	}

	storage := storage.NewStorage(us, config.SetUserInfoFunc)

	router := exampleop.SetupServer(issuer, storage)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	log.Default().Printf("listening at: %s", server.Addr)
	if config.TLS == nil {
		err = server.ListenAndServe()
	} else {
		err = server.ListenAndServeTLS(config.TLS.CertFile, config.TLS.KeyFile)
	}
	if err != nil {
		log.Fatal(err)
	}
	<-ctx.Done()
}
