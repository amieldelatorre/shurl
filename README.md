# Shurl
Shurl is a URL shortener written in go.

# Upcoming Features
- [x] Add postgres connection settings for fast cutover on failover
- [x] Add retries on db operations
- [x] Add idempotency key for create operations
- [ ] UI to create short urls
- [ ] Improve idempotency key by adding a hash for the request body to see if it is actually the same request
- [ ] Redis cache option
- [ ] User accounts
    - [ ] Require Login
    - [ ] Disable Registration
    - [ ] Allow anonymous creation
