# Shurl
Shurl is a URL shortener written in go.

## Generating a JWT Signing Key
```bash
openssl genpkey -algorithm EC -pkeyopt ec_paramgen_curve:secp521r1 -out jwt_private_key.pem
```

# Upcoming Features (maybe)
- [x] Add postgres connection settings for fast cutover on failover
- [x] Add retries on db operations
- [x] Add idempotency key for create operations
- [x] UI to create short urls
- [x] Improve idempotency key by adding a hash for the request body to see if it is actually the same request
- [x] Retry on the js fetch requests
- [x] User accounts
    - [x] Allow Login
    - [x] Allow Registration
    - [x] Allow anonymous creation
- [x] Background worker for deleting expired idempotency key
- [x] ~~Redis~~ Valkey cache option
- [x] Background worker for deleting expired short urls
- [ ] Make short url expiry configurable
