![Shock](http://github.com/jaredwilkening/Shock/raw/master/site/assets/img/shock_logo.png)
=====

About:
------

Shock is a platform to support computation, storage, and distribution. Designed from the ground up to be fast, scalable, fault tolerant, federated. 

Shock is RESTful. Accessible from desktops, HPC systems, exotic hardware, the cloud and your smartphone.

Shock is for scientific data. One of the challenges of large volume scientific data is that without often complex metadata it is of little to no value. Store and query(in development) complex metadata.   

Shock is an data management system. Annotate, anonymize, convert, filter, quality control, statically subsample at line speed bioinformatics sequence data. Extensible plug-in architecture(in development).

**Most importantly Shock is still very much in development. Be patient and contribute.**


Shock is actively being developed at [github.com/MG-RAST/Shock](http://github.com/MG-RAST/Shock).

<br>
To build:
---------

Unix/Macintosh 

Shock (requires go release.1 [golang.org/doc/install/source](http://golang.org/doc/install/source)):

    go get github.com/MG-RAST/Shock/...
  
To run (additional requires mongodb=>2.0.3):
  
    shock-server -conf <path_to_config_file>
    
The Shock configuration file is in INI file format. This file is documented in the shock.cfg.template file at the root level of the repository.

<br>
Command-line client:
-------------------

Alpha version available at [github.com/MG-RAST/ShockClient](http://github.com/MG-RAST/ShockClient).

<br>
Routes Overview
---------------
    
### Implemented API Routes (default port 8000):

#####GET

- [/](#get_slash)  resource listing
- [/node](#get_nodes)  list nodes, query
- [/node/{id}](#get_node)  view node, download file (full or partial)
- [/user](#get_users)  list users (admin users only)
- [/user/{id}](#get_user)  view user

#####PUT

- [/node/{id}](#put_node)  modify node, create index

#####POST
 
- [/node](#post_node)  create node
- [/user](#post_user)  create user

<br>
### Site Routes (default port 80):

    GET  /    
    GET  /raw    # listing of data dir
    GET  /assets # js, img, css, README.md

<br>
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
            "id": "4a6299ccb2cc44c2cd4b702cb98f2d9e", 
            "file": {
                "checksum": {
                    "md5": "05306fcb6f510ef7880863256797a486", 
                    "sha1": "10c97a28985623ca82cdf5337547406e7e48b1ed"
                }, 
                "name": "mgm4440286.3.json", 
                "size": 1861
            }, 
            "attributes": {
                "about": "metagenome", 
                "created": "2007-11-05 13:10:13", 
                "id": "mgm4440286.3", 
                "library": null, 
                "metadata": {
                    "biome-information_envo_lite": "animal-associated habitat", 
                    "external-ids_gold_id": "Gm00130", 
                    "external-ids_project_id": "28959, 28599", 
                    "external-ids_pubmed_id": "18698407, 18337718", 
                    "host-associated_age": "years/months/28/hh/mm/ss", 
                    "host-associated_body_site": "cecum", 
                    "host-associated_diet": "commercial chicken feed (Eagle Milling, AZ)", 
                    "host-associated_host_common_name": "Chicken", 
                    "host-associated_host_subject_id": "B", 
                    "host-associated_host_taxid": "9031", 
                    "host-associated_life_stage": "Adult", 
                    "host-associated_perturbation": "chicks were challenged via oral gavage with 1\u00d7105 CFU C. jejuni NCTC11168", 
                    "host-associated_samp_store_temp": "-80", 
                    "project-description_metagenome_name": "Chicken Cecum B Contigs", 
                    "sample-isolation-and-treatment_biomaterial_treatment": "DNA extraction", 
                    "sample-isolation-and-treatment_sample_isolation_description": "Fourteen days post challenge, birds from two pens (A&B) were euthanized and ceca collected for further analysis. Fresh cecal samples from two (C. jejuni-inoculated and C. jejuni-uninoculated) 28-day old chickens were analyzed. Cecal contents were collected using aseptic techniques. Samples were stored at &#8722;80\u00b0C until DNA extraction.", 
                    "sample-isolation-and-treatment_sample_isolation_reference": "18698407", 
                    "sample-origin_continent": "north_america", 
                    "sample-origin_country": "US", 
                    "sample-origin_geodetic_system": "wgs_84", 
                    "sample-origin_latitude": "40.1106", 
                    "sample-origin_location": "Urbana, IL", 
                    "sample-origin_longitude": "-88.2073", 
                    "sample-origin_sampling_timezone": "UTC", 
                    "sequencing_sequencing_center": "454 Life Sciences, Inc, Branford, CT", 
                    "sequencing_sequencing_method": "454"
                }, 
                "name": "Chicken Cecum B Contigs", 
                "project": "mgp101", 
                "sample": null, 
                "url": "http://api.metagenomics.anl.gov/metagenome/mgm4440286.3", 
                "version": 1
            }, 
            "acl": {
                "delete": [], 
                "read": [], 
                "write": []
            }, 
            "indexes": {}, 
            "version": "eeb8a92f954cc1691900497e537162fb", 
            "version_parts": {
                "acl_ver": "15251b8a2ba46ff4c6ce5baea5cd8b2a", 
                "attributes_ver": "8f0534ac82552d55420b82e7131f91f3", 
                "file_ver": "26c85f331645d98bf3f35d754c7ec352"
            }
        }, 
        "E": null, 
        "S": 200
    }

<br>
### User:

- uuid: unique identifier
- name: username
- passwd: all responds are masked "**********" 
- admin: boolean

##### user example:

    { 
        "D": {
            "uuid": "67394386a4acac62fdb851d78691ee48"
            "name": "joeuser", 
            "passwd": "**********", 
            "admin": false, 
        }, 
        "E": null, 
        "S": 200
    }

<br>
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

<br><br>
API
---

### Response wrapper:
All responses from Shock currently are in the following encoding. 

    {
        "D": <data in json or null>,
        "E": <error message or null>, 
        "S": <http status of request>
    }

<a name="get_slash"/>
<br>
### GET /

Description of resources available through this api

##### example
	
    curl -X GET http://<host>[:<port>]/
	
##### returns

    {"resources":["node", "user"],"url":"http://localhost:8000/","documentation":"http://localhost/","contact":"admin@host.com","id":"Shock","type":"Shock"}

<a name="post_node"/>
<br>
### POST /node

Create node

 - optionally takes user/password via Basic Auth. If set only that user with have access to the node
 - accepts multipart/form-data encoded 
 - to set attributes include file field named "attributes" containing a json file of attributes
 - to set file include file field named "file" containing any file

##### example
	
	curl -X POST [ --user user:password ] [ -F "attributes=@<path_to_json>" -F "file=@<path_to_data_file>" ] http://<host>[:<port>]/node
	
##### returns

    {
        "D": {<node>},
        "E": <error message or null>, 
        "S": <http status of response (also set in headers)>
    } 

<a name="get_nodes"/>
<br>
### GET /node

List nodes

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
	
	curl -X GET [ --user user:password ] http://<host>[:<port>]/node/[?skip=<skip>&limit=<count>][&query&<tag>=<value>]
		
##### returns

  	{
        "D": {[<array of nodes>]},
        "E": <error message or null>, 
        "S": <http status of request>
    }

<a name="get_node"/>
<br>
### GET /node/{id}

View node, download file (full or partial)

 - optionally takes user/password via Basic Auth
 - ?download - complete file download
 - ?download&index=size&part=1\[&part=2...\]\[chunksize=inbytes\] - download portion of the file via the size virtual index. Chunksize defaults to 1MB (1048576 bytes).

##### example	

	curl -X GET [ --user user:password ] http://<host>[:<port>]/node/{id}

##### returns

    {
        "D": {<node>},
        "E": <error message or null>, 
        "S": <http status of request>
    }

<a name="put_node"/>
<br>
### PUT /node/{id}

Modify node, create index

 - optionally takes user/password via Basic Auth
 
**Modify:** 

 - **Once the file or attributes of a node are set they are immutiable.**
 - accepts multipart/form-data encoded 
 - to set attributes include file field named "attributes" containing a json file of attributes
 - to set file include file field named "file" containing any file
 
##### example	
  
	curl -X PUT [ --user user:password ] [ -F "attributes=@<path_to_json>" -F "file=@<path_to_data_file>" ] http://<host>[:<port>]/node/{id}

  
##### returns

    {
        "D": {<node>},
        "E": <error message or null>, 
        "S": <http status of request>
    }

<br>
**Create index:**

 - currently available index types: size, record (for sequence file types)

##### example	

	curl -X PUT [ --user user:password ] http://<host>[:<port>]/node/{id}?index=<type>

##### returns

    {
        "D": null,
        "E": <error message or null>, 
        "S": <http status of request>
    }

<a name="post_user"/>
<br>
### POST /user

Create user

Requires Basic Auth encoded as 'username:password'. To create an admin user 'username:password:secret-key:true' where secret-key was specified at server start.
	
##### example	

    # regular user (when config Anonymous:create-user=true)
    curl -X POST --user joeuser:1234 http://<host>[:<port>]/user

    # regular user (when config Anonymous:create-user=false)
    curl -X POST --user joeuser:1234:supersupersecret:false http://<host>[:<port>]/user    

    # admin user
    curl -X POST --user joeuser:1234:supersupersecret:true http://<host>[:<port>]/user
	
##### returns

    {
        "D": {<user>},
        "E": <error message or null>, 
        "S": <http status of request>
    }

<a name="get_user"/>
<br>
### GET /user/{id}

View user

Requires Basic Auth encoded username:password. Regular user are able to see their own information while Admin user are able to access all. 

##### example	

    curl -X GET --user joeuser:1234 http://<host>[:<port>]/user/{id}

##### returns

    {
        "D": {<user>},
        "E": <error message or null>, 
        "S": <http status of request>
    }

<a name="get_users"/>
<br>
### GET /user

List users

Requires Basic Auth encoded username:password. Restricted to Admin users.

##### example	

    curl -X GET --user joeadmin:12345 http://<host>[:<port>]/user

##### returns

    {
        "D": {[<user>,...]},
        "E": <error message or null>, 
        "S": <http status of request>
    }

<br>
License
---

Copyright (c) 2010-2012, University of Chicago
All rights reserved.

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

