# Shock Setup for caching and data migration (new options with v2.0)

Author: Folker Meyer (folker@anl.gov)


## Caching

If started with a `--cache_path=/usr/local/shock/cache` or configured with a `PATH_CACHE` Shock will attempt download data present in Mongo from remote locations e.g. an S3 server. For this to work the Locations.yaml needs to provide the required details (URL, Auth, Bucket, Region, etc) *and* the node has to have a location entry pointing to said remote store.

It is useful to start Shock with `--expire_wait=1 ` to avoid having to wait for hours until the NodeReaper starts moving files, removing cached items or expiring files.

### Example:

This example assumes _no AUTH_, not something you want to try outside testing scenarios.

##### Preparation

Locations.yaml content:
~~~~
Locations:
 -  ID: "S3test1"
    Type: "S3"
    Description: "Example S3 Service "
    URL: "https://s3.example.com"
    AuthKey: "some_key"
    SecretKey: "another_key"
    Bucket: "mybucket1"
    Persistent: true
    Region: "us-east-1"
    Priority: 100
    MinPriority: 7
    Tier: 5
    Cost: 0
~~~~

Return of the `curl -X GET "localhost:7445/location/anls3_mgrast/info" | jq .`
~~~~
{
  "status": 200,
  "data": {
    "ID": "anls3_mgrast",
    "Description": "Argonne S3 Service bucket for MG-RAST",
    "type": "S3",
    "url": "https://s3.it.anl.gov:18082",
    "persistent": true,
    "priority": 0,
    "minpriority": 7,
    "tier": 5,
    "cost": 0,
    "bucket": "mgrast",
    "region": "us-east-1",
    "recoverycommand": ""
  },
  "error": null
}
~~~~

#### Configure the node to be aware of the remote location
setting a remote location for node 96576d58-6e2d-4bf5-8edf-8224cf29291c on a Shock server running on localhost
~~~~
curl -X POST -H 'Authorization: <secret>'  -H "Content-Type: application/json"  "localhost:7445/node/96576d58-6e2d-4bf5-8edf-8224cf29291c/locations/" -d '{"id":"S3test1" }
~~~~

#### Move data manually to remote location
Now you need to ensure that the file 96576d58-6e2d-4bf5-8edf-8224cf29291c.data exist on _S3test1_ and the credentials and config in Locations.yaml is correct. 
If present any indix files should be uploaded as a zipped archive named 96576d58-6e2d-4bf5-8edf-8224cf29291c.idx.zip to the same location.

We provide scripts for the node data (and index movement) that you can adapt to your needs.

#### Remove the local files and observe the Shock server download the _missing_ data seamlessly
The following command will download the _missing_ data item and create a local copy in PATH_CACHE (default: /usr/local/shock/cache)
~~~~
curl -X GET "localhost:7445/node/96576d58-6e2d-4bf5-8edf-8224cf29291c?download"
~~~~

#### Cache maintenance

Cached items (and their index files) are kept in the cache hierarchy for the duration specified by `cache_ttl` (default: 24 hours). The CacheReaper periodically scans the cache directory and removes files that have exceeded their TTL.

### Cache TTL

The `cache_ttl` configuration option controls how long cached files remain before they are eligible for eviction. It supports three time unit suffixes:

| Format | Meaning | Example |
|--------|---------|---------|
| `M` | Minutes | `30M` = 30 minutes |
| `H` | Hours | `24H` = 24 hours (default) |
| `D` | Days | `7D` = 7 days |

Set `cache_ttl` in the `[Cache]` section of the configuration file:

```ini
[Cache]
cache_path = /usr/local/shock/cache
cache_ttl = 24H
```

### Auto-Upload

When `auto_upload` is enabled, Shock automatically uploads newly created node files to a remote storage location. This is configured with the following options in the `[Cache]` section:

```ini
[Cache]
auto_upload = true
default_location = s3
upload_workers = 3
```

- `auto_upload`: Enable automatic upload of files to the default remote location.
- `default_location`: The Location ID (as defined in `Locations.yaml`) that receives the uploaded files.
- `upload_workers`: Number of concurrent upload workers (default: 3).

When a file is uploaded to Shock, the auto-upload workers will asynchronously copy it to the configured remote location. The node's `locations` field is updated once the upload completes.

### MinIO as S3 Backend

[MinIO](https://min.io/) is an S3-compatible object store that can be used as a local development and testing backend for Shock's S3 storage features. The repository includes a ready-to-use Docker Compose stack:

```bash
docker-compose -f docker-compose.minio.yml up -d shock-mongo shock-minio shock-minio-init shock-server
```

This starts:
- **MinIO** on port 9000 (API) and 9001 (console) with default credentials `minioadmin`/`minioadmin`
- **MongoDB** for metadata storage
- **Shock** configured with cache-to-S3 auto-upload pointing at MinIO

The MinIO Locations.yaml entry looks like:

```yaml
Locations:
  - ID: "s3"
    Type: "S3"
    Description: "Local MinIO S3"
    URL: "http://shock-minio:9000"
    AuthKey: "minioadmin"
    SecretKey: "minioadmin"
    Bucket: "shock-data"
    Persistent: true
    Region: "us-east-1"
    Priority: 100
    MinPriority: 0
    Tier: 5
    Cost: 0
```

See `test/config.d-minio/` for example configuration files used in integration testing.

## Data migration
Turn on _NODE_MIGRATION_ (e.g.  `--node_migration=true`) to create sets of files to be uploaded to remote locations.

Looking at the Location definition above (Locations.yaml):
~~~~
    Priority: 0
    MinPriority: 7
    Tier: 5
    Cost: 0
~~~~
Four settings determine the behavior of the node Migration.

Algorithm:
* From the locations with the highest `Priority` the lowest `Cost` location will be used first. If uploading failes, the next location will be tried.
* For each Node, the `MinPriority` value is checked to ensure no temporary files are moved to remote locations (unless desired). 
* The tier value describes the cost for staging the file back e.g. `Tier: 5 ` stores is slower than e.g. a `Tier: 3 ` store.



## Data file deletion
Turn on _NODE_DATA_REMOVAL_ (e.g. `--node_data_removal==true`).

If there are at least *MIN_REPLICA_COUNT* copies in the *Persistent* Locations, nodes (and their indices) can be removed from the the local disk. 
The NodeReaper will after expiring nodes that have reached their TTL and outputting nodes for migration, remove data matching the requirements above.

## Data migration

To migrate data a plug-in architecture is used (see `/scripts` in this repo). We provide a number of generic scripts but expect adopters to create their own/modify these scripts. 
The status of each node is (as usual) maintained in the Mongo database.

The location resource provides four calls to support a set of external migration tools. We provide tools for S3 and TSM at this time.

We note that externalizing the data migration we enabled massive scaling and allowed for the Shock server to remain lean.


### Server resources
#### `curl -X GET "localhost:7445/location/S3/info"`

will yield a JSON dump  information on the location itself
~~~~
{
  "status": 200,
  "data": {
    "ID": "S3",
    "Description": "Example S3 Service ",
    "type": "S3",
    "url": "https://s3.example.com",
    "persistent": true,
    "priority": 0,
    "minpriority": 7,
    "tier": 5,
    "cost": 0,
    "bucket": "mgrast",
    "region": "us-east-1",
    "recoverycommand": ""
  },
  "error": null
}
~~~~
#### `curl -X GET "localhost:7445/location/S3/missing`
This is the most important call for the data migration system. It lists all nodes that are eligible for migration to this resource (based on the priority and the resources minpriority).

The output will look something like the following:
~~~~
curl -s -X GET "localhost:7445/location/S3/missing" | jq .
{
  "status": 200,
  "data": [
    {
 
    {
      "id": "bc84d333-5158-43e5-b299-527128138e75",
      "version": "f29b9e1808fa5d3f05f4aea553ea35fa",
      "file": {
        "name": "",
        "size": 0,
        "checksum": {},
        "format": "",
        "virtual": false,
        "virtual_parts": [],
        "created_on": "0001-01-01T00:00:00Z",
        "locked": null
      },
      "attributes": null,
      "indexes": {},
      "version_parts": {
        "acl_ver": "b46701cc24139e5cca2100e09ec48c19",
        "attributes_ver": "2d7c3414972b950f3d6fa91b32e7920f",
        "file_ver": "804eb8573f5653088e50b14bbf4f634f",
        "indexes_ver": "99914b932bd37a50b983c5e7c90ae93b"
      },
      "tags": [],
      "linkage": [],
      "priority": 0,
      "created_on": "2019-09-06T16:46:01.548Z",
      "last_modified": "2019-09-06T17:18:25.508Z",
      "expiration": "0001-01-01T00:00:00Z",
      "type": "basic",
      "parts": null,
      "locations": [
        {
          "id": "S3"
          "stored: "false"
        }
      ]
    }
  ],
  "error": null
}
~~~~
The command `curl -s -X GET "localhost:7445/location/S3/missing" | jq .data[].id | tr -d \" ` will be useful as it returns just the IDs of nodes that need to be stored.

#### `curl -X GET "localhost:7445/location/S3/inflight`
This call will produce a list of all flights currently in flight, for a non batch system like S3 it would typically return an empty list. This is primarily intended for batched locations e.g. TSM.


#### `curl -X GET "localhost:7445/location/S3 /present`

This will list all nodes that have presently been stored on the S3 resource. We note that the primary purpose for this call is house cleaning. In the case of TSM this will generate a catalogue of files on tape.


### Scripts for data migration

### TSM
The TSM (batch mode) scripts require access to the DATA_PATH (note NOT the cache path) of the filesystem on the primary Shock server.

##### TSM Backup (`tsm_backup.sh`)
The script in `/scripts/tsm_backup.sh` will submit data to an already installed IBM Tivoli client (`dsmc`). It needs to be run with systems priviledges on a node with access to the file systems underlying the Shock data store and access to Tivoli.

The script will connect to Shock to retrieve list of ("missing") files to be moved to TSM. It will also connect to TSM to get list of files already in TSM. Once downloaded it will loop over the list of "missing" files and for each file in Shock list,
check if file is already in TSM (with `JSON{"id": "${LOCATION_NAME}", "stored": = "true" }` ). Files truly missing will be submitted via `dsmc` for backup and JSON to `/node/${id}/location/${LOCATION_NAME}/` with `{ "id": "${LOCATION_NAME}", "stored": "false" }`.


##### TSM Restore (`tsm_restore.sh`)
The script in `/scripts/tsm_restore.sh` will restore data from tape via the IBM Tivoli client (`dsmc`). 

This is intended to be either called directly by the Shock serve or called at defined intervals (once a day, once per hour) with a list of node IDs to be restored. It will restore <node_id>.data and the <node_id>/idx directory with its contents (if present).



#### S3, Azure, Shock and GoogleCloudStorage migration

~~~~
TBA by Andreas
~~~~

