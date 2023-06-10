# oidc-server

Dockerized OpenID Connect development server based on
https://github.com/zitadel/oidc/tree/main/example/server.
Storage and frontend can be adapted as needed for production.

# Setup

Authorization server listens on port `10001`. Expose accordingly.

# Runtime environment variables

`OIDC_ISSUER`: fully qualified domain name.

# Expected volume files

`/data/users.json`: key-value pairs of users indexed by username. See `users.example.json`.

