# Shock Setup for caching and data migration (new options with v2.0)

Author: Folker Meyer (folker@anl.gov)

## Concepts

- Traditionally Shock combines an on disk storage hierarchy with a Mongo database for metadata.
- As of version 1.0 a Shock server can download items present in its Mongo database from remote locations and cache them locally
- The "Traditional" behavior is the default model. The system is backward compatible.
- As of v1.0 Shock has become a hierarchical storage management system (HSM)
- A Shock server is made aware of remote locations via a Locations.yaml file, see the example
- Supported remote location types are: 
    * Shock (Type: Shock)
    * S3 (Type: S3)
    * IBM Tivoli TSM (Type: TSM)
- We hope to support the following in the future (pull requests welcome):
    * DAOS
    * Amazon Glacier
    * Google Cloud Storage
    * Microsoft Azure Data Storage
- Shock nodes are migrated listed for migration by external scripts (check the /scripts folder in this repo) by the server according to the paramters provided
- the new /location/<locname>/{missing|present|inflight} query lists the files that need to be backed up missing, the ones current being backed-up (inflight) and those stored already (present)


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

Cached items (and their index files) are kept in the cache hierarchy until for _cache_ttl_ hours (default=24).

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


## Misc

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

### Type.yaml from the configuration

~~~~
Types:
- ID: "default"
  Description: "default"
  Priority: 0
- ID: "temp"
  Description: "temporary file"
  Priority: 0
- ID: "EBI Submission Receipt"
  Description: "EBI Submission Receipt"
  Priority: 7
- ID: "VM"
  Description: "Virtual Machine"
  Priority: 1
- ID: "run-folder-archive-fastq"
  Description: "run-folder-archive-fastq"
  Priority: 9
  Data-Types:
    - fastq
- ID: "run-folder-archive-raw"
  Description: "run-folder-archive-raw"
  Priority: 4
- ID: "run-folder-archive-sav"
  Description: "run-folder-archive-sav"
  Priority: 9
  Data-Types:
    - sav
- ID: "run-folder-archive-thumbnails"
  Description: "run-folder-archive-thumbnails"
  Priority: 1
  Data-Types:
    - 
- ID: "inbox"
  Description: "MG-RAST inbox node"
  Priority: 1
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
- ID: "submission"
  Description: "MG-RAST submission"
  Priority: 9
- ID: "cv"
  Description: "Controlled Vocabulary"
  Priority: 7
- ID: "ontology"
  Description: "ontology"
  Priority: 7
- ID: "backup"
  Description: "Backup or Dump from another system e.g. MongoDB or MySQL"
  Priority: 9
- ID: "metadata"
  Description: "metadata"
  Priority: 7
- ID: "mixs"
  Description: "GSC MIxS Metadata file XLS format"
  Priority: 9
  Data-Types:
    - xls
    - xlsx
    - json
- ID: "reference"
  Description: "reference database"
  Priority: 7
- ID: "analysisObject"
  Description: "MG-RAST analysisObject"
  Priority: 1
- ID: "analysisRecipe"
  Description: "MG-RAST analysisRecipe"
  Priority: 1
- ID: "preference"
  Description: "MG-RAST user preference"
  Priority: 1

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