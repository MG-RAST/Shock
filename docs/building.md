# Building Shock

Shock is actively being developed at [github.com/MG-RAST/Shock](https://github.com/MG-RAST/Shock).

## Requirements

- **Go 1.22+**
- **MongoDB 3.6+**

## Building from Source

### Using the build script

The simplest way to build Shock is with the provided script, which automatically sets the version from git tags:

```bash
./compile-server.sh
```

### Using Go directly

Shock uses vendored dependencies. You must pass `-mod=vendor` to all Go commands:

```bash
CGO_ENABLED=0 go install -mod=vendor -installsuffix cgo -v \
  -ldflags="-X github.com/MG-RAST/Shock/shock-server/conf.VERSION=$(git describe --tags --long)" \
  ./shock-server/
```

### Vendored dependencies

All dependencies are checked into the `vendor/` directory and the build requires `-mod=vendor`. This ensures reproducible builds regardless of upstream module changes. Do not remove the `vendor/` directory.

## Docker

### Pull the pre-built image

```bash
docker pull mgrast/shock
```

### Build the Docker image locally

```bash
git clone https://github.com/MG-RAST/Shock.git
cd Shock
docker build -t mgrast/shock .
```

### Extract the binary from the Docker image

If you only need the statically compiled binary:

```bash
docker create --name shock mgrast/shock
mkdir -p bin
docker cp shock:/go/bin/shock-server ./bin/
docker rm shock
```

## MongoDB

Shock requires MongoDB for metadata storage. The easiest way to run MongoDB is with Docker Compose:

```bash
docker-compose up -d
```

This starts both Shock and MongoDB 3.6 using the included `docker-compose.yml`.

### Running MongoDB standalone

With Docker:

```bash
docker run --rm --name shock-mongo -p 27017:27017 mongo:3.6
```

Or install MongoDB via your system package manager and start it:

```bash
mongod --dbpath /data/db
```

## Configuration

The Shock configuration file uses INI format. A template is included at the root of the repository. See the [configuration guide](./configuration.md) for all available options.

## Running

```bash
shock-server -conf <path_to_config_file>
```

With Docker Compose (recommended):

```bash
docker-compose up -d
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

For S3 integration tests using MinIO:

```bash
docker-compose -f docker-compose.minio.yml build --no-cache
docker-compose -f docker-compose.minio.yml up -d shock-mongo shock-minio shock-minio-init shock-server
docker-compose -f docker-compose.minio.yml up --abort-on-container-exit shock-test
docker-compose -f docker-compose.minio.yml down -v
```

## Documentation

For further information about Shock's functionality, see the [documentation index](./README.md).
