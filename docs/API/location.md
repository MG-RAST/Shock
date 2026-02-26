# Location API Documentation

## API Routes for /location (default port 7445):

### OPTIONS

- Permitted by everyone:
  - all options requests respond with CORS headers and 200 OK

### POST

- Permitted by everyone:
  - N/A

- Permitted by admins:
  - `/location/<location_id>/info` - provide info on location
  - `/location/<location_id>/missing` - provide list of nodes missing in `<location_id>`
  - `/location/<location_id>/present` - provide list of nodes present in `<location_id>`
  - `/location/<location_id>/inflight` - provide list of nodes currently being transferred to `<location_id>`
  - `/location/<location_id>/restore` - provide list of nodes that need to be restored from `<location_id>`

### GET

- Permitted by everyone:
  - N/A

- Permitted by admins:
  - `/location/<location_id>/info` - provide info on location
  - `/location/<location_id>/missing` - provide list of nodes missing in `<location_id>`
  - `/location/<location_id>/present` - provide list of nodes present in `<location_id>`
  - `/location/<location_id>/inflight` - provide list of nodes currently being transferred to `<location_id>`
  - `/location/<location_id>/restore` - provide list of nodes that need to be restored from `<location_id>`

### PUT

- N/A

Note: Configuration is via the Locations.yaml file at server start time.

### DELETE

- N/A

## An overview of /location

### `GET /location/<location_id>/info`

Returns information about the specified location.

Example:
```
curl -X GET "localhost:7445/location/S3/info"
```

Response:
```json
{
  "status": 200,
  "data": {
    "ID": "S3",
    "Description": "Example S3 Service",
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
```

### `GET /location/<location_id>/missing`

This is the most important call for the data migration system. It lists all nodes that are eligible for migration to this resource (based on the priority and the resource's minpriority).

Example:
```
curl -X GET "localhost:7445/location/S3/missing"
```

Response:
```json
{
  "status": 200,
  "data": [
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
          "id": "S3",
          "stored": false
        }
      ]
    }
  ],
  "error": null
}
```

To get just the IDs of nodes that need to be stored:
```
curl -s -X GET "localhost:7445/location/S3/missing" | jq .data[].id | tr -d \"
```

### `GET /location/<location_id>/inflight`

This call will produce a list of all nodes currently in flight (being transferred). For a non-batch system like S3, it would typically return an empty list. This is primarily intended for batched locations like TSM.

Example:
```
curl -X GET "localhost:7445/location/S3/inflight"
```

### `GET /location/<location_id>/present`

This will list all nodes that have been stored on the specified resource. The primary purpose for this call is housekeeping. In the case of TSM, this will generate a catalog of files on tape.

Example:
```
curl -X GET "localhost:7445/location/S3/present"
```

### `GET /location/<location_id>/restore`

This call will list all nodes that need to be restored from the specified location.

Example:
```
curl -X GET "localhost:7445/location/S3/restore"
```

## Scripts for data migration

### TSM Backup

The script in `/scripts/tsm_backup.sh` will submit data to an already installed IBM Tivoli client (`dsmc`). It needs to be run with system privileges on a node with access to the file systems underlying the Shock data store and access to Tivoli.

The script will connect to Shock to retrieve a list of ("missing") files to be moved to TSM. It will also connect to TSM to get a list of files already in TSM. Once downloaded, it will loop over the list of "missing" files and for each file in the Shock list, check if the file is already in TSM (with `JSON{"id": "${LOCATION_NAME}", "stored": = "true" }`). Files truly missing will be submitted via `dsmc` for backup and JSON to `/node/${id}/location/${LOCATION_NAME}/` with `{ "id": "${LOCATION_NAME}", "stored": "false" }`.

### TSM Restore

The script in `/scripts/tsm_restore.sh` will restore data from tape via the IBM Tivoli client (`dsmc`).

This is intended to be either called directly by the Shock server or called at defined intervals (once a day, once per hour) with a list of node IDs to be restored. It will restore `<node_id>.data` and the `<node_id>/idx` directory with its contents (if present).

### S3 migration

Scripts for S3, Azure, Shock, and Google Cloud Storage migration are available in the `/scripts` directory.

## API by example

All examples use curl but can be easily modified for any HTTP client/library.
**Note**: Authentication is required for most of these commands.

### Query location

Retrieve info for ${LOCATION_ID}:
```
curl -s -X POST -H "$AUTH" "${SHOCK_SERVER_URL}/location/${LOCATION_ID}/info"
```

Retrieve a list of nodes missing at ${LOCATION_ID} (this will respect the MinPriority setting for ${LOCATION_ID}):
```
curl -s -X POST -H "$AUTH" "${SHOCK_SERVER_URL}/location/${LOCATION_ID}/missing"
```

## Response wrapper

All responses from Shock currently use the following encoding:

```json
{
  "data": <JSON or null>,
  "error": <string or null: error message>,
  "status": <int: http status code>,
  "limit": <int: paginated requests only>, 
  "offset": <int: paginated requests only>,
  "total_count": <int: paginated requests only>
}
```

## Configuring locations

See the [Configuration Guide](../configuration.md#locations-configuration) for detailed information on configuring locations.
