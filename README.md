![Shock](http://github.com/jaredwilkening/Shock/raw/master/misc/shock_logo.png)
=====

About:
------

Shock is a platform for computation, storage, and distribution. Designed from the ground up to be fast, scalable, fault tolerant, federated. 

Shock is RESTful. Accessible from desktops, HPC systems, exotic hardware, the cloud and your smartphone.

Shock is for scientific data. One of the challenges of large volume scientific data is that without often complex metadata it is of little to no value. Store and query(in development) complex metadata.   

Shock is an active storage layer. Annotate, anonymize, convert, filter, quality control, statically subsample at line speed bioinformatics sequence data. Extensible plug-in architecture(in development).

**Most importantly Shock is still very much in development. Be patient and contribute.**

Road Map:
---------

Coming soon

To build:
---------

Unix/Macintosh 

To install go weekly.2012-01-20 ([weekly.golang.org/doc/install/source](http://weekly.golang.org/doc/install/source)):
    
    hg clone -u release https://code.google.com/p/go
    hg pull
    hg update weekly.2012-01-20
    cd go/src
    ./all.bash
    <add ../bin to $PATH>

To build Shock:

    git clone <this repo>
    cd Shock
    export GOPATH=`pwd`
    go install shock/shock-server
  
To run (additional requires mongodb=>2.0.3):
  
    ./bin/shock-server -port=<port#> -dataroot=<path_to_data_root> -mongo=<mongo_host(s)>
  
Data Types
----------

### Node:

##### id
unique identifier

##### file 

 - file name 
 - file size
 - file checksum(s) 

##### attributes
arbitrary json

##### acl
access control (in development)

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

 - to set attributes include file field named "attributes" containing a json file of attributes
 - to set file include file field named "file" containing any file

##### example
	
	curl -X POST [ -F "attributes=@<path_to_json>" -F "file=@<path_to_data_file>" ] http://<shock_host>[:<port>]/node
	
##### returns

    {
        "C":"",
        "D": {<node>},
        "E": <error message or null>, 
        "S": <http status of request>
    } 

<br/>
### List nodes:
GET /node

 - by adding ?offset=N you get the nodes starting at N+1 
 - by adding ?limit=N you get a maximum of N nodes returned 

Querying:<br/> 
All attributes are queriable. For example if a node has in it's attributes "about" : "metagenome" the url /node/?query&about=metagenome would return it and all other nodes with that attribute. Address of nested attributes like "metadata": { "env_biome": "ENVO:human-associated habitat", ... } is done via a dot notation /node/?query&metadata.env_biome=ENVO:human-associated habitat.

##### example
	
	curl -X GET http://<shock_host>[:<port>]/node/[?offset=<offset>&limit=<count>][&query&<tag>=<value>]
		
##### returns

  	{
        "C":"",
        "D": {[<array of nodes>]},
        "E": <error message or null>, 
        "S": <http status of request>
    }

<br/>	
### Get node:
GET /node/:nodeid
	
 - ?download - complete file download
	
##### example	

	curl -X GET http://<shock_host>[:<port>]/node/:nodeid
	
##### returns

    {
        "C":"",
        "D": {<node>},
        "E": <error message or null>, 
        "S": <http status of request>
    }

