

### API Routes for /location (default port 7445):

(TODO: move this into openapi) 

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


## Configuring locations


### Example `Locations.yaml` file
This is a copy of the contents of Example_Locations.yaml file from the repo. Please check that file as well for updates.
~~~~
Locations:
 -  ID: "S3"
    Type: "S3"
    Description: "Example S3 Service "
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
 -  ID: "S3SSD"
    Type: "S3"
    Description: "Example_S3_SSD Service"
    URL: "https://s3-ssd.example.com"
    AuthKey: "yet_another_key"
    SecretKey: "yet_another_nother_key"
    Bucket: "ssd"
    Persistent: true
    Region: "us-east-1" Priority:
    Priority: 0
    Tier: 3
    Cost: 0
 -  ID: "shock"
    Type: "shock"
    Description: "shock service"
    URL: "shock.example.org"
    AuthKey: ""
    SecretKey: ""
    Prefix: ""
    Priority: 0
    Tier: 5
    Cost: 0
 -  ID: "tsm"
    Type: "tsm_archive"  
    Description: "archive service"
    URL: ""
    AuthKey: ""
    SecretKey: ""
    Prefix:  ""
    Restorecommand: "dsmc restore %ID%  -latest"
    Priority: 0
    Tier: 10
    Cost: 0
~~~~


### Complete config from the source code
~~~~
// Location set of storage locations
type LocationConfig struct {
	ID          string `bson:"ID" json:"ID" yaml:"ID" `                           // e.g. ANLs3 or local for local store
	Description string `bson:"Description" json:"Description" yaml:"Description"` // e.g. ANL official S3 service
	Type        string `bson:"type" json:"type" yaml:"Type" `                     // e.g. S3
	URL         string `bson:"url" json:"url" yaml:"URL"`                         // e.g. http://s3api.invalid.org/download&id=
	Token       string `bson:"token" json:"-" yaml:"Token" `                      // e.g.  Key or password
	Prefix      string `bson:"prefix" json:"-" yaml:"Prefix"`                     // e.g. any prefix needed
	AuthKey     string `bson:"AuthKey" json:"-" yaml:"AuthKey"`                   // e.g. AWS auth-key
	Persistent  bool   `bson:"persistent" json:"persistent" yaml:"Persistent"`    // e.g. is this a valid long term storage location
	Priority    int    `bson:"priority" json:"priority" yaml:"Priority"`          // e.g. location priority for pushing files upstream to this location, 0 is lowest, 100 highest
	MinPriority int    `bson:"minpriority" json:"minpriority" yaml:"MinPriority"` // e.g. minimum node priority level for this location (e.g. some stores will only handle non temporary files or high value files)
	Tier        int    `bson:"tier" json:"tier" yaml:"Tier"`                      // e.g. class or tier 0= cache, 3=ssd based backend, 5=disk based backend, 10=tape archive
	Cost        int    `bson:"cost" json:"cost" yaml:"Cost"`                      // e.g. cost per GB for this store, default=0

	S3Location  `bson:",inline" json:",inline" yaml:",inline"` // extensions specific to S3
	TSMLocation `bson:",inline" json:",inline" yaml:",inline"` // extension sspecific to IBM TSM
}

// S3Location S3 specific fields
type S3Location struct {
	Bucket    string `bson:"bucket" json:"bucket" yaml:"Bucket" `
	Region    string `bson:"region" json:"region" yaml:"Region" `
	SecretKey string `bson:"SecretKey" json:"-" yaml:"SecretKey" ` // e.g.g AWS secret-key
}

// TSMLocation IBM TSM specific fields
type TSMLocation struct {
	Recoverycommand string `bson:"recoverycommand" json:"recoverycommand" yaml:"Recoverycommand" `
}
~~~~




