# Shock Project Scratchpad

## Current Task: Integration & Acceptance Tests

### Status: Complete - All Tests Passing

### What was done

Implemented the full integration/acceptance test suite for the Shock API:

**Files created:**
- `test/shock-server-test.conf` - INI config for test Shock server (basic auth, port 7445, mongo at shock-mongo-test)
- `shock-server/integration/helpers_test.go` - TestMain, HTTP helpers, user seeding, node creation utilities
- `shock-server/integration/node_crud_test.go` - 9 CRUD tests (create, read, update, delete, auth checks)
- `shock-server/integration/node_query_test.go` - 4 query/pagination tests
- `shock-server/integration/node_acl_test.go` - 8 ACL management tests
- `shock-server/integration/node_upload_test.go` - 7 file upload/download tests
- `shock-server/integration/preauth_test.go` - 3 pre-auth download tests
- `shock-server/router/router_test.go` - 8 router/middleware tests (CORS, not-found, root, options)

**Files modified:**
- `docker-compose.test.yml` - Added `shock-server-test` service, updated `shock-test` with entrypoint override
- `Dockerfile` - Fixed Python installation for newer Alpine
- `Dockerfile.test` - Removed broken `ENTRYPOINT ["/bin/sh", "-c"]`

### Test counts
- Integration tests: 31 test functions + TestMain
- Router tests: 8 test functions (run locally, no Docker needed)
- Total new tests: 39

### Running tests
- Router tests (local): `go test -v ./shock-server/router/`
- Integration tests (Docker): start services then `docker-compose run --rm --entrypoint "" shock-test go test -v ./shock-server/integration/...`
- All new tests (Docker): `docker-compose -f docker-compose.test.yml up --abort-on-container-exit`

### Architecture
```
shock-mongo-test (MongoDB 3.6)
  ← shock-server-test (compiled binary, port 7445, basic auth)
    ← shock-test-runner (go test client, HTTP requests)
```

### Key decisions
- Used `assert` (already vendored) instead of `require` to avoid vendor directory issues
- Helper functions use `t.Fatal` for critical failures
- Each test creates/cleans up its own nodes via `t.Cleanup`
- TestMain seeds users via direct MongoDB and drops database on teardown
- Router tests use httptest (no Docker needed)

### Issues found and fixed during Docker verification
1. INI section `[MongoDB]` → `[Mongodb]` (case-sensitive parser)
2. Missing `[External] api-url` caused preauth URLs to use `localhost` instead of Docker hostname
3. Missing `[Other] force_yes=true` could cause interactive prompts
4. `ENTRYPOINT ["/bin/sh", "-c"]` in Dockerfile.test broke docker-compose command passing
5. Complex inline shell script in docker-compose replaced with exec-form command
