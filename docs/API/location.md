

### API Routes for /location (default port 7445):

##### OPTIONS

- Permitted by everyone:
  - all options requests respond with CORS headers and 200 OK

##### POST
- Permitted by everyone:
- N/A

- Permitted by admis 
  - `/location/<location_id>/info`  provide info on location
  - `/location/<location_id>/missing`  provide list of nodes missing in <location_id>
  - `/location/<location_id>/present`  provide list of nodes present in <location_id>
  - `/location/<location_id>/inflight`  provide list of nodes currently being transferred to <location_id>
  
##### GET

- Permitted by everyone:
 - N/A

- Permitted by admis 
  - `/location/<location_id>/info`  provide info on location
  - `/location/<location_id>/missing`  provide list of nodes missing in <location_id>
  - `/location/<location_id>/present`  provide list of nodes present in <location_id>
  - `/location/<location_id>/inflight`  provide list of nodes currently being transferred to <location_id>
  
  


##### PUT

- N/A

Note: Configurations is via the Locations.yaml file at server start time.


##### DELETE

- N/A

<br>


## An overview of /location

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

#### `curl -X GET "localhost:7445/location/S3/inflight`
This call will produce a list of all flights currently in flight, for a non batch system like S3 it would typically return an empty list. This is primarily intended for batched locations e.g. TSM.


#### `curl -X GET "localhost:7445/location/S3 /present`

This will list all nodes that have presently been stored on the S3 resource. We note that the primary purpose for this call is house cleaning. In the case of TSM this will generate a catalogue of files on tape.


### Scripts for data migration

#### TSM Backup
The script in `/scripts/tsm_backup.sh` will submit data to an already installed IBM Tivoli client (`dsmc`). It needs to be run with systems priviledges on a node with access to the file systems underlying the Shock data store and access to Tivoli.

The script will connect to Shock to retrieve list of ("missing") files to be moved to TSM. It will also connect to TSM to get list of files already in TSM. Once downloaded it will loop over the list of "missing" files and for each file in Shock list,
check if file is already in TSM (with `JSON{"id": "${LOCATION_NAME}", "stored": = "true" }` ). Files truly missing will be submitted via `dsmc` for backup and JSON to `/node/${id}/location/${LOCATION_NAME}/` with `{ "id": "${LOCATION_NAME}", "stored": "false" }`.

#### S3 migration

~~~~
TBA by Andreas
~~~~



API by example
-----------------
All examples use curl but can be easily modified for any http client / library. 
__Note__: Authentication is required for most of these commands
<br>
#### query location 

# retrieve info for ${LOCATION_ID}
curl -s -X POST -H "$AUTH" "${SHOCK_SERVER_URL}/location/${LOCATION_ID}/missing" 

# retrieve a list of nodes missing at ${LOCATION_ID} (this will respect the MinPriority setting for ${LOCATION_ID})
curl -s -X POST -H "$AUTH" "${SHOCK_SERVER_URL}/location/${LOCATION_ID}/info"




API
---

### Response wrapper:

All responses from Shock currently are in the following encoding. 

  {
    "data": <JSON or null>,
    "error": <string or null: error message>,
    "status": <int: http status code>,
    "limit": <int: paginated requests only>, 
    "offset": <int: paginated requests only>,
    "total_count": <int: paginated requests only>
  }

<br>

### GET /

Description of resources available through this api

##### example
	
    curl -X GET http://<host>[:<port>]/
	
##### returns

    {"resources":["node"],"url":"http://localhost:7445/","documentation":"http://localhost:7445/documentation.html","contact":"admin@host.com","id":"Shock","type":"Shock"}

<br>







