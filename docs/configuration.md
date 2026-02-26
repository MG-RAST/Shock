# Shock Configuration Guide

This document provides a comprehensive guide to configuring the Shock server, including all configuration options, file formats, and command-line arguments.

## Table of Contents

1. [Overview](#overview)
2. [Configuration File](#configuration-file)
3. [Command-line Arguments](#command-line-arguments)
4. [Locations Configuration](#locations-configuration)
5. [Types Configuration](#types-configuration)
6. [Data Migration and Caching](#data-migration-and-caching)
7. [Restore Functionality](#restore-functionality)
8. [Examples](#examples)

## Overview

The Shock configuration system consists of several components:

1. **Main Configuration File**: An INI-format file (typically `shock-server.conf`) that contains the core server settings
2. **Locations.yaml**: Defines storage locations for data migration and caching
3. **Types.yaml**: Defines node types and their priorities
4. **Command-line Arguments**: Override settings in the configuration files

Configuration files are typically located in the `/etc/shock.d/` directory, but can be specified with the `-conf` command-line argument.

## Configuration File

The main configuration file uses INI format with sections and key-value pairs. Below are the available sections and options:

### [Admin]
| Option | Type | Default | Description |
|--------|------|---------|-------------|
| email | string | "" | Administrator email address |
| users | string | "" | Comma-separated list of admin users |

### [Anonymous]
| Option | Type | Default | Description |
|--------|------|---------|-------------|
| read | bool | true | Allow anonymous read access |
| write | bool | true | Allow anonymous write access |
| delete | bool | true | Allow anonymous delete access |

### [Address]
| Option | Type | Default | Description |
|--------|------|---------|-------------|
| api-ip | string | "0.0.0.0" | IP address to bind the API server |
| api-port | int | 7445 | Port for the API server |

### [External]
| Option | Type | Default | Description |
|--------|------|---------|-------------|
| api-url | string | "http://localhost" | External URL for the API |

### [Auth]
| Option | Type | Default | Description |
|--------|------|---------|-------------|
| basic | bool | false | Enable basic authentication |
| globus_token_url | string | "" | Globus token URL for authentication |
| globus_profile_url | string | "" | Globus profile URL for authentication |
| oauth_urls | string | "" | Comma-separated list of OAuth URLs |
| oauth_bearers | string | "" | Comma-separated list of OAuth bearers |
| cache_timeout | int | 60 | Authentication cache timeout in minutes |
| use_auth | bool | true | Enable authentication (disable for debugging) |

### [Runtime]
| Option | Type | Default | Description |
|--------|------|---------|-------------|
| expire_wait | int | 60 | Wait time for reaper in minutes |
| GOMAXPROCS | string | "" | Number of CPU cores to use (empty uses Go default) |
| max_revisions | int | 3 | Maximum number of node revisions to keep (values < 0 mean keep all) |

### [Log]
| Option | Type | Default | Description |
|--------|------|---------|-------------|
| perf_log | bool | false | Enable performance logging |
| rotate | bool | true | Enable log rotation |
| logoutput | string | "both" | Log output destination: "console", "file", or "both" |
| trace | bool | false | Enable trace logging |
| debuglevel | int | 0 | Debug level (0-3) |

### [Mongodb]
| Option | Type | Default | Description |
|--------|------|---------|-------------|
| attribute_indexes | string | "" | Comma-separated list of attribute indexes |
| database | string | "ShockDB" | MongoDB database name |
| hosts | string | "mongo" | MongoDB host(s) |
| password | string | "" | MongoDB password |
| user | string | "" | MongoDB username |

### [Node-Indices]
Custom node indices can be defined in this section. Each index can have the following options:
- unique: true/false
- dropDups: true/false
- sparse: true/false

Example:
```
[Node-Indices]
name=unique:true,dropDups:true,sparse:false
```

### [Paths]
| Option | Type | Default | Description |
|--------|------|---------|-------------|
| site | string | "/usr/local/shock/site" | Path to site files |
| data | string | "/usr/local/shock/data" | Path to data files |
| logs | string | "/var/log/shock" | Path to log files |
| local_paths | string | "/var/tmp" | Path to local temporary files |
| pidfile | string | "" | Path to PID file |

### [Cache]
| Option | Type | Default | Description |
|--------|------|---------|-------------|
| cache_path | string | "" | Path to cache directory. If set, the system will function as a cache |

### [Migrate]
| Option | Type | Default | Description |
|--------|------|---------|-------------|
| min_replica_count | int | 2 | Minimum number of locations required before enabling local Node file deletion |
| node_migration | bool | false | Enable node migration to remote locations |
| node_data_removal | bool | false | Enable removal of data for nodes with at least MIN_REPLICA_COUNT copies |

### [SSL]
| Option | Type | Default | Description |
|--------|------|---------|-------------|
| enable | bool | false | Enable SSL |
| key | string | "" | Path to SSL key file |
| cert | string | "" | Path to SSL certificate file |

### [Other]
| Option | Type | Default | Description |
|--------|------|---------|-------------|
| reload | string | "" | Path or URL to shock data (WARNING: this will drop all current data) |
| conf | string | "shock-server.conf" | Path to config file |
| no_config | bool | false | Do not use config file |
| force_yes | bool | false | Force yes to all prompts |
| version | bool | false | Show version |
| fullhelp | bool | false | Show detailed usage without "--" prefixes |
| help | bool | false | Show usage |
| debug_auth | bool | false | Enable more detailed reasons for rejected auth (for debugging) |

## Command-line Arguments

All configuration options can be overridden with command-line arguments. The format is:

```bash
shock-server --option=value
```

For example:

```bash
shock-server --conf=/path/to/shock-server.conf --api-port=8080
```

Common command-line arguments:

- `--conf`: Path to the configuration file
- `--no_config`: Do not use a configuration file
- `--api-port`: Port for the API server
- `--api-ip`: IP address to bind the API server
- `--node_migration`: Enable node migration to remote locations
- `--node_data_removal`: Enable removal of data for nodes with at least MIN_REPLICA_COUNT copies
- `--min_replica_count`: Minimum number of locations required before enabling local Node file deletion
- `--cache_path`: Path to cache directory
- `--expire_wait`: Wait time for reaper in minutes

## Locations Configuration

The Locations.yaml file defines storage locations for data migration and caching. It is located in the same directory as the main configuration file.

### Format

```yaml
Locations:
  - ID: "location_id"
    Type: "location_type"
    Description: "description"
    URL: "url"
    AuthKey: "auth_key"
    SecretKey: "secret_key"
    Bucket: "bucket_name"
    Persistent: true/false
    Region: "region"
    Priority: priority_value
    MinPriority: min_priority_value
    Tier: tier_value
    Cost: cost_value
    # Additional type-specific fields
```

### Common Fields

| Field | Description |
|-------|-------------|
| ID | Unique identifier for the location |
| Type | Type of storage location (S3, Shock, TSM, etc.) |
| Description | Human-readable description |
| URL | URL for the storage location |
| AuthKey | Authentication key |
| SecretKey | Secret key for authentication |
| Persistent | Whether this is a valid long-term storage location |
| Priority | Location priority for pushing files upstream (0 is lowest, 100 highest) |
| MinPriority | Minimum node priority level for this location |
| Tier | Storage tier (0=cache, 3=SSD, 5=disk, 10=tape archive) |
| Cost | Cost per GB for this store (default=0) |

### Type-Specific Fields

#### S3 Location
```yaml
Bucket: "bucket_name"
Region: "region"
```

#### Azure Location
```yaml
Account: "account_name"
Container: "container_name"
```

#### Google Cloud Location
```yaml
Project: "project_name"
```

#### IRods Location
```yaml
Zone: "zone"
User: "user"
Password: "password"
Hostname: "hostname"
Port: port_number
```

#### Glacier Location
```yaml
Vault: "vault_name"
```

### Example Locations.yaml

```yaml
Locations:
  - ID: "S3"
    Type: "S3"
    Description: "Example S3 Service"
    URL: "https://s3.example.com"
    AuthKey: "some_key"
    SecretKey: "another_key"
    Bucket: "mybucket1"
    Persistent: true
    Region: "us-east-1"
    Priority: 0
    Tier: 5
    Cost: 0
    MinPriority: 7
  - ID: "S3SSD"
    Type: "S3"
    Description: "Example_S3_SSD Service"
    URL: "https://s3-ssd.example.com"
    AuthKey: "yet_another_key"
    SecretKey: "yet_another_nother_key"
    Bucket: "ssd"
    Persistent: true
    Region: "us-east-1"
    Priority: 0
    Tier: 3
    Cost: 0
  - ID: "shock"
    Type: "shock"
    Description: "shock service"
    URL: "shock.example.org"
    AuthKey: ""
    SecretKey: ""
    Prefix: ""
    Priority: 0
    Tier: 5
    Cost: 0
  - ID: "tsm"
    Type: "tsm_archive"
    Description: "archive service"
    URL: ""
    AuthKey: ""
    SecretKey: ""
    Prefix: ""
    Recoverycommand: "dsmc restore %ID% -latest"
    Priority: 0
    Tier: 10
    Cost: 0
```

## Types Configuration

The Types.yaml file defines node types and their priorities. It is located in the same directory as the main configuration file.

### Format

```yaml
Types:
  - ID: "type_id"
    Description: "description"
    Priority: priority_value
    Data-Types:
      - data_type1
      - data_type2
```

### Fields

| Field | Description |
|-------|-------------|
| ID | Unique identifier for the type |
| Description | Human-readable description |
| Priority | Priority value for the type (0 is lowest, higher values indicate higher priority) |
| Data-Types | List of data types associated with this type |

### Example Types.yaml

```yaml
Types:
  - ID: "default"
    Description: "default"
    Priority: 0
  - ID: "temp"
    Description: "temporary file"
    Priority: 0
  - ID: "metagenome"
    Description: "MG-RAST metagenome"
    Priority: 9
    Data-Types:
      - fa
      - fasta
      - fastq
      - fq
      - bam
      - sam
  - ID: "image"
    Description: "image file"
    Priority: 1
    Data-Types:
      - jpeg
      - jpg
      - gif
      - tif
      - png
```

## Data Migration and Caching

Shock supports data migration to remote locations and caching of data from remote locations.

### Data Migration

Data migration is controlled by the following configuration options:

- `node_migration`: Enable node migration to remote locations
- `node_data_removal`: Enable removal of data for nodes with at least MIN_REPLICA_COUNT copies
- `min_replica_count`: Minimum number of locations required before enabling local Node file deletion

When `node_migration` is enabled, Shock will attempt to migrate data to remote locations defined in Locations.yaml. The migration process is based on the following algorithm:

1. From the locations with the highest `Priority`, the lowest `Cost` location will be used first
2. For each Node, the `MinPriority` value is checked to ensure no temporary files are moved to remote locations
3. The `Tier` value describes the cost for staging the file back (lower tier values are faster)

### Caching

Caching is controlled by the `cache_path` configuration option. If this option is set, Shock will function as a cache and attempt to download nodes present in MongoDB that are not present on local disk from one of the configured Locations.

When a node is requested and not found locally, Shock will:

1. Check if the node exists in MongoDB
2. If it does, check if it has a location entry pointing to a remote location
3. Download the node data from the remote location
4. Store it in the `cache_path` directory
5. Serve the data to the client

Cached items are kept in the cache hierarchy for a configurable time period (default is 24 hours).

## Restore Functionality

Shock supports restoring data from archive locations like tape storage. This is controlled by the following node properties:

- `Restore`: Boolean flag indicating whether a node has been marked for restoring from an external location

When a node is marked for restore, Shock will attempt to retrieve it from the archive location. This is particularly useful for tape-based storage systems like IBM Tivoli Storage Manager (TSM).

### Restore Process

1. A node is marked for restore using the `SetRestore()` method
2. External scripts (like `tsm_restore.sh`) are used to retrieve the data from the archive location
3. Once the data is restored, the `UnSetRestore()` method is called to indicate that the restore has been completed

## Examples

### Basic Configuration

```ini
[Admin]
email = admin@example.com
users = admin1,admin2

[Address]
api-ip = 0.0.0.0
api-port = 7445

[Mongodb]
hosts = localhost
database = ShockDB

[Paths]
site = /usr/local/shock/site
data = /usr/local/shock/data
logs = /var/log/shock
```

### Enabling Data Migration

```ini
[Migrate]
min_replica_count = 2
node_migration = true
node_data_removal = true
```

### Enabling Caching

```ini
[Cache]
cache_path = /usr/local/shock/cache
```

### Running Shock Server

To run the Shock server with a specific configuration file:

```bash
shock-server -conf /path/to/shock-server.conf
```

With Docker:

```bash
mkdir -p /mnt/shock-server/log
mkdir -p /mnt/shock-server/data
export DATADIR="/mnt/shock-server"
docker run --rm --name shock-server -p 7445:7445 \
  -v ${DATADIR}/shock-server.cfg:/shock-config/shock-server.cfg \
  -v ${DATADIR}/log:/var/log/shock \
  -v ${DATADIR}/data:/usr/local/shock \
  --link=shock-server-mongodb:mongodb \
  mgrast/shock /go/bin/shock-server --conf /shock-config/shock-server.cfg
```

### Data Migration Example

To enable data migration with a short expiration wait time:

```bash
shock-server --conf=/path/to/shock-server.conf --node_migration=true --expire_wait=1
```

This will start the Shock server with data migration enabled and set the expiration wait time to 1 minute, which is useful for testing as it avoids having to wait for hours until the NodeReaper starts moving files.
