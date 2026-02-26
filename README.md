# Shock

Shock is a platform to support computation, storage, and distribution. Designed from the ground up to be fast, scalable, fault tolerant, federated. (see [Shock: Active Storage for Multicloud Streaming Data Analysis](http://ieeexplore.ieee.org/abstract/document/7406331/), Big Data Computing (BDC), 2015 IEEE/ACM 2nd International Symposium on, 2015)

Shock is RESTful. Accessible from desktops, HPC systems, exotic hardware, the cloud and your smartphone.

Shock is for scientific data. One of the challenges of large volume scientific data is that without often complex metadata it is of little to no value. Shock allows storage and querying of complex metadata.

Shock is a data management system. The long term goals of Shock include the ability to annotate, anonymize, convert, filter, perform quality control, and statically subsample at line speed bioinformatics sequence data. Extensible plug-in architecture is in development.

Shock is actively being developed at [github.com/MG-RAST/Shock](https://github.com/MG-RAST/Shock).

## Quick Start (Docker Compose)

The fastest way to try Shock is with Docker Compose:

```bash
docker-compose up -d
curl http://localhost:7445/
```

This starts Shock and MongoDB. See the [tutorial](docs/tutorial.md) for a hands-on walkthrough.

## Requirements

- **Go 1.22+** (for building from source)
- **MongoDB 3.6+** (metadata storage)

## Building

### From source

Use the provided build script, which sets the version from git tags:

```bash
./compile-server.sh
```

Or build directly with Go (note: vendored dependencies require `-mod=vendor`):

```bash
CGO_ENABLED=0 go install -mod=vendor -installsuffix cgo -v \
  -ldflags="-X github.com/MG-RAST/Shock/shock-server/conf.VERSION=$(git describe --tags --long)" \
  ./shock-server/
```

### Docker

Pull the pre-built image:

```bash
docker pull mgrast/shock
```

Or build it locally:

```bash
git clone https://github.com/MG-RAST/Shock.git
cd Shock
docker build -t mgrast/shock .
```

To extract just the statically compiled binary from the image:

```bash
VERSION=$(docker run --rm mgrast/shock shock-server --version | grep version | grep -o v[0-9].* | tr -d '\n')
echo $VERSION
docker create --name shock mgrast/shock
mkdir -p bin
docker cp shock:/go/bin/shock-server ./bin/shock-server-${VERSION}
docker rm shock
```

## Configuration

The Shock configuration file uses INI format. A template is included at the root of the repository. Storage locations and node types are configured via `Locations.yaml` and `Types.yaml` respectively.

See the [configuration guide](docs/configuration.md) for all options.

## Running

```bash
shock-server -conf <path_to_config_file>
```

With Docker Compose (recommended):

```bash
docker-compose up -d
```

The default `docker-compose.yml` starts both Shock and MongoDB. For S3-compatible storage with MinIO, use the MinIO stack:

```bash
docker-compose -f docker-compose.minio.yml up -d shock-mongo shock-minio shock-minio-init shock-server
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

For S3 integration tests with MinIO:

```bash
docker-compose -f docker-compose.minio.yml up --abort-on-container-exit shock-test
```

## API Overview

| Route | Description |
|-------|-------------|
| `POST /node` | Create a new node (with or without file upload) |
| `GET /node/{nid}` | Retrieve node metadata |
| `GET /node/{nid}?download` | Download node file |
| `PUT /node/{nid}` | Update node attributes |
| `DELETE /node/{nid}` | Delete a node |
| `GET/PUT /node/{nid}/acl/` | View or modify access control |
| `PUT /node/{nid}/index/{type}` | Create a file index |
| `GET/POST /node/{nid}/locations/` | Manage storage locations |
| `GET /node?query&key=value` | Query nodes by metadata |

See the [API documentation](docs/API/api.md) for full details and examples.

## Documentation

- [Getting Started Tutorial](docs/tutorial.md)
- [Building Shock](docs/building.md)
- [Configuration Guide](docs/configuration.md)
- [Concepts](docs/concepts.md)
- [API Reference](docs/API/api.md)
- [Caching and Data Migration](docs/caching_and_data_migration.md)
- [Data Types](docs/Data-Types.md)
- [Use Cases](docs/use_cases.md)

## License

BSD-3-Clause
