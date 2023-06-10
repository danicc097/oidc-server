package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/danicc097/openapi-go-gin-postgres-sqlc/external/oidc-server/exampleop"
	"github.com/danicc097/openapi-go-gin-postgres-sqlc/external/oidc-server/storage"
)

func main() {
	ctx := context.Background()

	issuer := os.Getenv("OIDC_ISSUER")
	port := "10001" // for internal network

	usersFile, err := os.Open("/data/users.json")
	if err != nil {
		log.Fatal(err)
	}
	defer usersFile.Close()

	us, err := storage.NewUserStore(issuer, usersFile)
	if err != nil {
		log.Fatal("could not create user store: ", err)
	}

	storage := storage.NewStorage(us)

	router := exampleop.SetupServer(issuer, storage)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	log.Default().Printf("listening at: %s", server.Addr)
	err = server.ListenAndServe()
	// if running in localhost manually add certs, else let traefik handle https
	// err := server.ListenAndServeTLS("certificates/localhost.pem", "certificates/localhost-key.pem")
	if err != nil {
		log.Fatal(err)
	}
	<-ctx.Done()
}
