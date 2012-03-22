![Shock](http://github.com/jaredwilkening/Shock/raw/master/misc/shock_logo.png)
=====

About:
------

Shock is a platform for computation, storage, and distribution. Designed from the ground up to be fast, scalable, fault tolerant, federated. 

Shock is RESTful. Accessible from desktops, HPC systems, exotic hardware, the cloud and your smartphone.

Shock is for scientific data. One of the challenges of large volume scientific data is that without often complex metadata it is of little to no value. Store and query(in development) complex metadata.   

Shock is an active storage layer. Annotate, anonymize, convert, filter, quality control, statically subsample at line speed bioinformatics sequence data. Extensible plug-in architecture(in development).

**Most importantly Shock is still very much in development. Be patient and contribute.**

To build:
---------

Unix/Macintosh 

To install go weekly.2012-03-13 ([weekly.golang.org/doc/install/source](http://weekly.golang.org/doc/install/source)):
    
    hg clone -u release https://code.google.com/p/go
    hg pull
    hg update weekly.2012-03-13
    cd go/src
    ./all.bash
    <add ../bin to $PATH>

To build Shock:

    go get github.com/MG-RAST/Shock/shock-server
  
To run (additional requires mongodb=>2.0.3):
  
    shock-server -port=<port to listen on> \
                 -data=<data directory to store on disk files> \
                 -mongo=<hostname(s) of mongodb> \
                 -secretkey=<secret key>

Command-line client:
-------------------

Alpha version available @ [github.com/MG-RAST/ShockClient](http://github.com/MG-RAST/ShockClient)  

Data Types
----------

### Node:

- id: unique identifier
- file: name, size, checksum(s).
- attributes: arbitrary json. Queriable.
- acl: arrays of user uuids corresponding to read, write, delete access controls

##### node example (metagenome from MG-RAST):

    {
        "D": {
            "id": "75ccb7e590c8fc8df90f3759847e8947", 
            "file": {
                "checksum": {}, 
                "name": "", 
                "size": 0
            }, 
            "attributes": {
                "about": "metagenome", 
                "created": "2011-05-19 11:08:48", 
                "id": "mgm4456668.3", 
                "library": "mgl1812", 
                "metadata": {
                    "ANONYMIZED_NAME": "sample499", 
                    "COMMON_NAME": "M14Fcsw", 
                    "DESCRIPTION": "Bacterial Community Variation in Human Body Habitats Across Space and Time", 
                    "TAXON_ID": "9606", 
                    "TITLE": "Bacterial Community Variation in Human Body Habitats Across Space and Time", 
                    "altitude": "0.0", 
                    "anatomical_sample_site": "FMA:Feces", 
                    "assigned_from_geo": "n", 
                    "biological_specimen": "M14Fcsw", 
                    "body_habitat": "UBERON:feces", 
                    "body_site": "UBERON:feces", 
                    "collection_date": "2008-2009", 
                    "common_sample_site": "stool", 
                    "country": "GAZ:United States of America", 
                    "depth": "0", 
                    "elevation": "1624.097656", 
                    "env_biome": "ENVO:human-associated habitat", 
                    "env_feature": "ENVO:human-associated habitat", 
                    "env_matter": "ENVO:human-associated habitat", 
                    "host_individual": "M1", 
                    "latitude": "40.0149856", 
                    "longitude": "-105.2705456", 
                    "original_sample_site": "stool", 
                    "public": "y", 
                    "samp_collect_device": "swab with sterile saline", 
                    "samp_size": "1 swab", 
                    "sample_id": "qiime:145415", 
                    "sample_name": "M14Fcsw", 
                    "sex": "male", 
                    "study_id": "qiime:449"
                }, 
                "name": "1812", 
                "project": "mgp81", 
                "sample": "mgs1812", 
                "url": "http://api.metagenomics.anl.gov/metagenome/mgm4456668.3", 
                "version": 1
            }, 
            "indexes": {},
            "acl": {
                "delete": [], 
                "read": [], 
                "write": []
            } 
        }, 
        "E": null, 
        "S": 200
    }

### User:

- uuid: unique identifier
- name: username
- passwd: all responds are masked "**********" 
- admin: boolean

##### user example:

    {
        "C": "", 
        "D": {
            "uuid": "67394386a4acac62fdb851d78691ee48"
            "name": "joeuser", 
            "passwd": "**********", 
            "admin": false, 
        }, 
        "E": null, 
        "S": 200
    }

### Index:

Currently there is support for two types of indices: virtual and file. 

##### virtual index:

A virtual index is one that can be generated on the fly without support of precalculated data. The current working example of this 
is the size virtual index. Based on the file size and desired chunksize the partitions become individually addressable. 

##### file index:

Currently in early development the file index is a json file stored on disk in the node's directory.  

    # abstract form
    {
    	index_type : <type>,
    	filename : <filename>,
    	checksum_type : <type>,
    	version : <version>,
    	index : [
    		[<position>,<length>,<optional_checksum>]...
    	]
    }
    
    # example
    {
    	"index_type" : "fasta",
    	"filename" : "none",
    	"checksum_type" : "none",
    	"version" : 1,
    	"index" : [[0,1861]]
    }
    
API
---

### Response wrapper:
All responses from Shock currently are in the following encoding. 

    {
        "C":"",
        "D": <data in json or null>,
        "E": <error message or null>, 
        "S": <http status of request>
    }

### Create node:
POST /node (multipart/form-data encoded)

 - optionally takes user/password via Basic Auth. If set only that user with have access to the node
 - to set attributes include file field named "attributes" containing a json file of attributes
 - to set file include file field named "file" containing any file

##### example
	
	curl -X POST [ --user user:password ] [ -F "attributes=@<path_to_json>" -F "file=@<path_to_data_file>" ] http://<shock_host>[:<port>]/node
	
##### returns

    {
        "C":"",
        "D": {<node>},
        "E": <error message or null>, 
        "S": <http status of request>
    } 

### List nodes:
GET /node

 - optionally takes user/password via Basic Auth. Grants access to non-public data
 - by adding ?skip=N you get the nodes starting at N+1 
 - by adding ?limit=N you get a maximum of N nodes returned 

##### querying
All attributes are queriable. For example if a node has in it's attributes "about" : "metagenome" the url 

    /node/?query&about=metagenome
    
would return it and all other nodes with that attribute. Address of nested attributes like "metadata": { "env_biome": "ENVO:human-associated habitat", ... } is done via a dot notation 

    /node/?query&metadata.env_biome=ENVO:human-associated%20habitat

Multiple attributes can be selected in a single query and are treated as AND operations

    /node/?query&metadata.env_biome=ENVO:human-associated%20habitat&about=metagenome
    
**Note:** all special characters like a space must be url encoded.

##### example
	
	curl -X GET [ --user user:password ] http://<shock_host>[:<port>]/node/[?skip=<skip>&limit=<count>][&query&<tag>=<value>]
		
##### returns

  	{
        "C":"",
        "D": {[<array of nodes>]},
        "E": <error message or null>, 
        "S": <http status of request>
    }

### Get node:
GET /node/:nodeid

 - optionally takes user/password via Basic Auth
 - ?download - complete file download
 - ?download&index=size&part=1\[&part=2...\]\[chunksize=inbytes\] - download portion of the file via the size virtual index. Chunksize defaults to 1MB (1048576 bytes).

##### example	

	curl -X GET [ --user user:password ] http://<shock_host>[:<port>]/node/:nodeid

##### returns

    {
        "C":"",
        "D": {<node>},
        "E": <error message or null>, 
        "S": <http status of request>
    }

### Create user:
POST /user

Requires Basic Auth encoded username:password. To create an admin user include :secret_key specified at server start.
	
##### example	

    # regular user 
    curl -X POST --user joeuser:1234 http://<shock_host>[:<port>]/user
    
    # admin user
    curl -X POST --user joeuser:1234:supersupersecret http://<shock_host>[:<port>]/user
	
##### returns

    {
        "C":"",
        "D": {<user>},
        "E": <error message or null>, 
        "S": <http status of request>
    }

### Get user:
GET /user/:uuid

Requires Basic Auth encoded username:password. Regular user are able to see their own information while Admin user are able to access all. 

##### example	

    curl -X GET --user joeuser:1234 http://<shock_host>[:<port>]/user/:uuid

##### returns

    {
        "C":"",
        "D": {<user>},
        "E": <error message or null>, 
        "S": <http status of request>
    }

### List users:
GET /user

Requires Basic Auth encoded username:password. Restricted to Admin users.

##### example	

    curl -X GET --user joeadmin:12345 http://<shock_host>[:<port>]/user

##### returns

    {
        "C":"",
        "D": {[<user>,...]},
        "E": <error message or null>, 
        "S": <http status of request>
    }

