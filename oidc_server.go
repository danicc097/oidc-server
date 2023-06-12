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
type Config[T storage.User] struct {
	// SetUserInfoFunc overrides population of userinfo based on scope.
	// Example:
	// func SetUserInfoFunc(user *CustomUser, userInfo *oidc.UserInfo, scope string, clientID string) {
	// 	switch scope {
	// 	case oidc.ScopeOpenID:
	// 		userInfo.Subject = user.ID
	// 	case oidc.ScopeEmail:
	// 		userInfo.Email = user.Email
	// 		userInfo.EmailVerified = oidc.Bool(user.EmailVerified)
	// 	case oidc.ScopeProfile:
	// 		userInfo.PreferredUsername = user.Username
	// 		userInfo.Name = user.FirstName + " " + user.LastName
	// 		userInfo.FamilyName = user.LastName
	// 		userInfo.GivenName = user.FirstName
	// 		userInfo.Locale = oidc.NewLocale(user.PreferredLanguage)
	// 	case oidc.ScopePhone:
	// 		userInfo.PhoneNumber = user.Phone
	// 		userInfo.PhoneNumberVerified = user.PhoneVerified
	// 	case AuthScope:
	// 		userInfo.AppendClaims(AuthClaim, map[string]interface{}{
	// 			"is_admin": user.IsAdmin,
	// 		})
	// 	case CustomScope:
	// 		userInfo.AppendClaims(CustomClaim, customClaim(clientID))
	// 	}
	// }
	SetUserInfoFunc storage.SetUserInfoFunc[T]
	// TLS runs the server with the given certificate.
	TLS *struct {
		CertFile string
		KeyFile  string
	}
}

// Runs starts the OIDC server.
func Run[T storage.User](config Config[T]) {
	ctx := context.Background()

	issuer := os.Getenv("ISSUER")
	port := "10001" // for internal network
	usersDataDir := path.Join(os.Getenv("DATA_DIR"), "users")

	us, err := storage.NewUserStore[T](issuer, usersDataDir)
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
