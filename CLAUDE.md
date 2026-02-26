# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Shock is a RESTful, high-performance object storage platform for scientific data. It provides file storage with MongoDB-backed metadata, multi-cloud integration (S3, Azure, GCS, IBM TSM), and in-storage operations for bioinformatics data processing.

## Build Commands

```bash
# Build the server (sets version from git tags)
./compile-server.sh

# Or directly with Go
CGO_ENABLED=0 go install -installsuffix cgo -v -ldflags="-X github.com/MG-RAST/Shock/shock-server/conf.VERSION=$(git describe --tags --long)" ./shock-server/

# Docker build
docker build -t mgrast/shock .
```

## Running

```bash
shock-server -conf <path_to_config_file>
```

## Testing

Tests run in Docker with MongoDB:

```bash
# Run all tests
./run-tests.sh all

# Run tests for a specific package
./run-tests.sh package ./shock-server/node

# Generate coverage report
./run-tests.sh coverage

# Clean test environment
./run-tests.sh clean
```

Test utilities are available in `shock-server/test/` with mock implementations for database and filesystem.

## Architecture

### Entry Point
- `shock-server/main.go` - Server initialization and HTTP routing via chi

### Core Packages

| Package | Purpose |
|---------|---------|
| `node/` | Core data model - node CRUD, file management, indexing, ACLs, locations |
| `file/` | File handling and format-specific processors (FASTA, SAM, etc.) |
| `auth/` | Authentication mechanisms (basic, globus, oauth) |
| `controller/node/` | REST API endpoint handlers |
| `conf/` | Configuration management (INI + YAML files) |
| `db/` | MongoDB connectivity |
| `location/` | Storage location management for multi-cloud |
| `user/` | User management |

### Data Flow
```
REST API → Controllers → Node Package → MongoDB (metadata) + Filesystem/Cloud (data)
```

### Configuration Files
- Main config: INI format (e.g., `shock-server.conf`)
- Storage locations: `Locations.yaml`
- Node types: `Types.yaml`

### Key API Routes
- `POST/GET/PUT/DELETE /node/{nid}` - Node operations
- `/node/{nid}/acl/` - Access control
- `/node/{nid}/index/{type}` - File indexing
- `/node/{nid}/locations/` - Storage locations
- `/preauth/{id}` - Pre-authenticated access

## Dependencies

- **Go 1.22**
- **MongoDB** - Required for metadata storage
- **gopkg.in/mgo.v2** - MongoDB driver
- **github.com/stretchr/testify** - Testing assertions
- AWS, Azure, and GCS SDKs for cloud storage

## Profiling

Built-in pprof available on port 6060 when running.
