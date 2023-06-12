# oidc-server

Dockerized OpenID Connect development server based on
https://github.com/zitadel/oidc/tree/main/example/server.
Storage and frontend can be adapted as needed for production.

# Setup

Authorization server listens on port `10001`. Expose accordingly.

# Runtime environment variables

`OIDC_ISSUER`: fully qualified domain name.

# Expected volume files

- `/data/users/*.json`: JSON files with key-value pairs of users for easier
  testing. Keys are ignored. Server will shutdown if duplicated IDs are
  found. The `/data/users` folder is watched for changes. See
  `storage/user.go`'s `User` for available fields.

- `/data/redirect_uris.txt`: valid redirect URIs to load at startup.
