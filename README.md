# oidc-server

Dockerized OpenID Connect development server based on
https://github.com/zitadel/oidc/tree/main/example/server.
Storage and frontend can be adapted as needed for production.

# Setup

Authorization server listens on port `10001`. Expose accordingly.

# Runtime environment variables

`OIDC_ISSUER`: fully qualified domain name.

# Expected volume files

`/data/users/*.json`: JSON files with key-value pairs of users indexed by
username for easier testing. Users will be combined in ascending
filename order and overriden by username. See `storage/user.go`'s `User`.
`/redirect_uris.txt`: valid redirect URIs to load at startup.
