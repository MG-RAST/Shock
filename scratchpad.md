# Shock Project Scratchpad

## Current State: On `feature/s3-setup` branch

- **Branch**: `feature/s3-setup`
- **Previous**: PR #402 merged to master, tag v2.1.0-alpha

## Session Summary (2026-02-26) — S3 Integration Tests

### Files Created
1. `shock-server/integration/s3_location_test.go` — 10 new integration tests for S3/location endpoints

### Files Modified
2. `shock-server/integration/helpers_test.go` — Added `nodeLocation` struct, `locationConfigData` struct, `Locations` field to `nodeData`, 4 helper functions (`parseLocationConfigData`, `parseLocationList`, `parseSingleLocation`, `waitForLocation`)

### Tests Added (10 total)
- **Config**: `TestLocationS3ConfigInfo`, `TestLocationInfoRequiresAdmin`
- **Auto-upload**: `TestAutoUploadCreatesLocation`, `TestAutoUploadMultipleFiles`
- **CRUD**: `TestViewNodeLocations`, `TestViewSpecificLocation`, `TestAdminDeleteLocation`, `TestDeleteLocationRequiresAdmin`
- **Edge cases**: `TestEmptyNodeHasNoLocations`, `TestLocationNotFoundOnNode`

### Key decisions
- Only `testify/assert` is vendored (not `require`); used `if !assert.X() { return }` pattern instead
- `waitForLocation` uses exponential backoff (500ms→4s cap, 30s total timeout)
- `nodeLocation.RequestedDate` is `*string` to avoid fragile time parsing
- Skipped `/location/s3/present` (has bson bug: string "true" vs bool true)

## Previous Session — MinIO Docker Compose Setup

### Files Created
1. `docker-compose.minio.yml` — Full compose stack: MinIO + MongoDB + Shock server + test runner
2. `test/shock-server-minio.conf` — Shock config with `[Cache]` section (TTL=5M, auto_upload=true, default_location=s3)
3. `test/config.d-minio/Locations.yaml` — S3 location pointing to MinIO (minioadmin/minioadmin, bucket=shock-data)
4. `test/config.d-minio/Types.yaml` — Copied from test/config.d/Types.yaml

### Files Modified
5. `Dockerfile` — Added boto-s3-*.py to /usr/local/bin/ + created /var/cache/shock

### Architecture
- MinIO as S3 backend (port 9000 API, 9001 console)
- minio-init one-shot creates `shock-data` bucket
- mongo-seed creates admin/user1 accounts in shock_minio_test DB
- Shock server reads config from /etc/shock.d/ (conf + Locations.yaml + Types.yaml mounted individually)
- Test runner uses SHOCK_SERVER_URL=http://shock-server:7445

### Key Design Decisions
- CONFIG_PATH = dir of config file → Locations.yaml/Types.yaml must be in same dir as .conf
- Volume mounts: individual file mounts into /etc/shock.d/ rather than directory mount
- Uses modern MINIO_ROOT_USER/MINIO_ROOT_PASSWORD (not deprecated MINIO_ACCESS_KEY)
- Healthchecks on all stateful services; depends_on with conditions for proper startup ordering
- **Shared config**: test runner reads server config via `SHOCK_CONFIG` env var using `goconfig` library (single source of truth for MongoDB host/database)
- Backward compatible: falls back to `MONGO_HOST` env var + hardcoded defaults when `SHOCK_CONFIG` is not set
- No additional services needed: auth is built-in basic, metadata is MongoDB-only, S3 via boto3 scripts

### Verification Steps
```bash
docker-compose -f docker-compose.minio.yml build --no-cache
docker-compose -f docker-compose.minio.yml up -d shock-mongo shock-minio shock-minio-init shock-server
curl http://localhost:7445/
docker-compose -f docker-compose.minio.yml up --abort-on-container-exit shock-test
docker-compose -f docker-compose.minio.yml down -v
```

## Open Issues

| Issue | Priority | Description |
|-------|----------|-------------|
| [#403](https://github.com/MG-RAST/Shock/issues/403) | Medium | anonymize.Reader: test-specific logic + goroutine race conditions |
| [#404](https://github.com/MG-RAST/Shock/issues/404) | Low | QueueUpload race: called before node persisted to DB |
| [#405](https://github.com/MG-RAST/Shock/issues/405) | Low | 3 remaining `go vet` warnings (oauth struct tag, 2 unreachable code) |
| [#406](https://github.com/MG-RAST/Shock/issues/406) | Medium | 9 packages missing test files |

## Key Technical Notes

- **Always build with `-mod=vendor`** — golib and tinyftp vendored APIs differ from upstream
- **Go version**: 1.22 (per go.mod)
- **Router**: Chi (go-chi/chi/v5)
- **Test infra**: Docker Compose with MongoDB 3.6, test server on port 7445, basic auth
- **2 packages require Docker**: `integration` (needs Shock server), `node` (needs MongoDB)
