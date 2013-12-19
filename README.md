![Shock](https://raw.github.com/MG-RAST/Shock/master/shock-server/site/assets/img/shock_logo.png)
=====

About:
------

Shock is a platform to support computation, storage, and distribution. Designed from the ground up to be fast, scalable, fault tolerant, federated. 

Shock is RESTful. Accessible from desktops, HPC systems, exotic hardware, the cloud and your smartphone.

Shock is for scientific data. One of the challenges of large volume scientific data is that without often complex metadata it is of little to no value. Shock allows storage and querying of complex metadata.   

Shock is a data management system. The long term goals of Shock include the ability to annotate, anonymize, convert, filter, perform quality control, and statically subsample at line speed bioinformatics sequence data. Extensible plug-in architecture is in development.

**Most importantly Shock is still very much in development. Be patient and contribute.**

Shock is actively being developed at [github.com/MG-RAST/Shock](http://github.com/MG-RAST/Shock).

Building:
---------
Shock (requires mongodb=>2.0.3, go=>1.1.0 [golang.org](http://golang.org/), git, mercurial, bazaar):

    go get github.com/MG-RAST/Shock/...

Built binary will be located in env configured $GOPATH or $GOROOT depending on the Go configuration.

Configuration:
--------------
The Shock configuration file is in INI file format. There is a template of the config file located at the root level of the repository.

Running:
--------------
To run:
  
    shock-server -conf <path_to_config_file>

Routes Overview
---------------
    
### API/Site Routes (default port 7445):

#####OPTIONS

- all options request respond with CORS headers and 200 OK

#####GET

- [/](#get_slash)  resource listing
- [/assets]() js, img, css, README.md
- [/documentation.html]() this documentation

- [/node](#get_nodes)  list nodes, query
- [/node/{id}](#get_node)  view node, download file (full or partial)
- [/node/{id}/acl]()  view node acls
- [/node/{id}/acl/{type}]()  view node acls of type {type}

#####PUT

- [/node/{id}](#put_node)  modify node
- [/node/{id}/acl]()  modify node acls
- [/node/{id}/acl/{type}]()  modify node acls of type {type}
- [/node/{id}/index/{type}]()  create node indexes

#####POST
 
- [/node](#post_node)  create node

#####DELETE

- [/node/{id}]()  delete node

<br>

Authentication:
---------------
Shock supports multiple forms of Authentication via plugin modules. Credentials are cached for 1 hour to speed up high transaction loads. Server restarts will clear the credential cache.

### Globus Online 
In this configuration Shock locally stores only uuids for users that it has already seen. The registration of new users is done exclusively with the external auth provider. The user api is disabled in this mode.

Examples:

    # globus online username & password
    curl --user username:password ...

    # globus online bearer token 
    curl -H "Authorization: OAuth $TOKEN" ...


<br>

Data Types
----------

### Node:

- id: unique identifier
- file: name, size, checksum(s).
- attributes: arbitrary json. Queriable.
- indexes: A set of indexes to use.
- version: a version stamp for this node.

##### node example:
    
    {
        "data": {
            "attributes": null, 
            "file": {
                "checksum": {}, 
                "format": "", 
                "name": "", 
                "size": 0, 
                "virtual": false, 
                "virtual_parts": []
            }, 
            "id": "130cadb5-9435-4bd9-be13-715ec40b2bb5", 
            "relatives": [], 
            "type": [], 
            "version": "4da883924aa8ae9eb95f6cd247f2f554"
        }, 
        "error": null, 
        "status": 200
    }

<br>
### Index:

Currently available index types include: size (virtual, does not require index creation), line, chunkrecord and record (for sequence file types), and bai (bam index)

##### virtual index:

A virtual index is one that can be generated on the fly without support of precalculated data. The current working example of this 
is the size virtual index. Based on the file size and desired chunksize the partitions become individually addressable. 

##### bam index (bai):

To use the bam index feature, the <a href="http://samtools.sourceforge.net/">SAMtools</a> package must be installed on the machine that is running the Shock server with the samtools executable in the path of the user that is running the Shock server.

Also, in order to use this feature, you must sort your .bam file using the 'samtools sort' command before uploading the file into Shock.

<br><br>

API by example
-----------------
All examples use curl but can be easily modified for any http client / library. 
__Note__: Authentication is required for most of these commands
<br>
#### Node Creation ([details](#post_node)):

    # without file or attributes
    curl -X POST http://<host>[:<port>]/node

    # with attributes
    curl -X POST -F "attributes=@<path_to_json_file>" http://<host>[:<port>]/node
    
    # with file, using multipart form
    curl -X POST -F "upload=@<path_to_data_file>" http://<host>[:<port>]/node

    # with file, without using multipart form
    curl -X POST --data-binary @<path_to_data_file> http://<host>[:<port>]/node
        (note: Posting an empty file in this way will result in an empty node with no file rather than an empty node with an empty file)

    # copying data file from another node
    curl -X POST -F "copy_data=<copy_node_id>" http://<host>[:<port>]/node

    # with file local to the shock server
    curl -X POST -F "path=<path_to_data_file>" -F "action=<action_type>" http://<host>[:<port>]/node
        (note: The action_type is one of keep_file (node points to file path given), copy_file (file is copied to shock data directory), or move_file (file is moved to shock data directory).  The move_file action only works if user running Shock has permissions to move the file.)
    
    # with file upload in N parts (part uploads may be done in parallel and out of order)
    curl -X POST -F "parts=N" http://<host>[:<port>]/node
    curl -X PUT -F "1=@<file_part_1>" http://<host>[:<port>]/node/<node_id>
    curl -X PUT -F "2=@<file_part_2>" http://<host>[:<port>]/node/<node_id>
    ...
    curl -X PUT -F "N=@<file_part_N>" http://<host>[:<port>]/node/<node_id>
	
    # with file upload in N parts where N is unknown at node creation time (part uploads may be done in parallel and out of order)
    curl -X POST -F "parts=unknown" http://<host>[:<port>]/node
    curl -X PUT -F "1=@<file_part_1>" http://<host>[:<port>]/node/<node_id>
    curl -X PUT -F "2=@<file_part_2>" http://<host>[:<port>]/node/<node_id>
    ...
    curl -X PUT -F "parts=close" http://<host>[:<port>]/node/<node_id>

<br>
#### Node retrieval ([details](#get_node)):

    # node information
    curl -X GET http://<host>[:<port>]/node/{id}

    # download file
    curl -X GET http://<host>[:<port>]/node/{id}/?download

    # download first 1mb of file
    curl -X GET http://<host>[:<port>]/node/{id}/?download&index=size&part=1
        
    # download first 10mb of file
    curl -X GET http://<host>[:<port>]/node/{id}/?download&index=size&chunk_size=10485760&part=1

    # download Nth 10mb of file
    curl -X GET http://<host>[:<port>]/node/{id}/?download&index=size&chunk_size=10485760&part=N

    # download entire bam file in human readable sam alignments
    curl -X GET http://<host>[:<port>]/node/{id}/?download&index=bai

    # download bam alignments overlapped with specified region (ref_id:start_pos-end_pos)
    curl -X GET http://<host>[:<port>]/node/{id}/?download&index=bai&region=chr1:1-20000

    # download bam alignments with selected arguments supported by "samtools view"
    curl -X GET http://<host>[:<port>]/node/{id}/?download/index=bai&head&headonly&count&flag=[INT]&lib=[STR]&mapq=[INT]&readgroup=[STR]
    (note: All the arguments are optional and can be used with or without the region, but the index=bai is required)
    
<br>
#### Node acls: 

    # view all acls
    curl -X GET http://<host>[:<port>]/node/{id}/acl/

    # view specific acls
    curl -X GET http://<host>[:<port>]/node/{id}/acl/[ all | read | write | delete | owner ]

    # changing owner (chown)
    curl -X PUT http://<host>[:<port>]/node/{id}/acl/owner?users=<user-id_or_uuid>

    # adding user to all acls (except owner)
    curl -X PUT http://<host>[:<port>]/node/{id}/acl/all?users=<user-ids_or_uuids>

    # adding user to specific acls
    curl -X PUT http://<host>[:<port>]/node/{id}/acl/[ read | write | delete ]?users=<user-ids_or_uuids>
    
    # deleting user from all acls (except owner)
    curl -X DELETE http://<host>[:<port>]/node/{id}/acl/all?users=<user-ids_or_uuids>    
    
    # deleting user to specific acls
    curl -X DELETE http://<host>[:<port>]/node/{id}/acl/[ read | write | delete ]?users=<user-ids_or_uuids>
    
<br>
#### Querying ([details](#get_nodes)):

    # by attribute key value
    curl -X GET http://<host>[:<port>]/node/?query&<key>=<value>

    # by attribute key value, limit 10
    curl -X GET http://<host>[:<port>]/node/?query&<key>=<value>&limit=10

    # by attribute key value, limit 10, offset 10
    curl -X GET http://<host>[:<port>]/node/?query&<key>=<value>&limit=10&offset=10

<br>

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

<a name="get_slash"/>
<br>
### GET /

Description of resources available through this api

##### example
	
    curl -X GET http://<host>[:<port>]/
	
##### returns

    {"resources":["node"],"url":"http://localhost:7445/","documentation":"http://localhost:7445/documentation.html","contact":"admin@host.com","id":"Shock","type":"Shock"}

<a name="post_node"/>
<br>
### POST /node

Create node

 - optionally takes user/password via Basic Auth. If set only that user with have access to the node
 - accepts multipart/form-data encoded 
 - to set attributes include file field named "attributes" containing a json file of attributes
 - to set file include file field named "upload" containing any file **or** include field named "path" containing the file system path to the file accessible from the Shock server

##### example
	
	curl -X POST [ see Authentication ] [ -F "attributes=@<path_to_json>" ( -F "upload=@<path_to_data_file>" || -F "path=<path_to_file>") ] http://<host>[:<port>]/node
	
##### returns

    {
        "data": {<node>},
        "error": <error message or null>, 
        "status": <http status of response (also set in headers)>
    } 

<a name="get_nodes"/>
<br>
### GET /node

List nodes

 - optionally takes user/password via Basic Auth. Grants access to non-public data
 - by adding ?offset=N you get the nodes starting at N+1 
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
	
	curl -X GET [ see Authentication ] http://<host>[:<port>]/node/[?offset=<offset>&limit=<count>][&query&<tag>=<value>]
		
##### returns

    {
      "data": {[<array of nodes>]},
      "error": <string or null: error message>,
      "status": <int: http status code>,
      "limit": <limit>, 
      "offset": <offset>,
      "total_count": <count>
    }

<a name="get_node"/>
<br>
### GET /node/{id}

View node, download file (full or partial)

 - optionally takes user/password via Basic Auth
 - ?download - complete file download
 - ?download&index=size&part=1\[&part=2...\]\[chunksize=inbytes\] - download portion of the file via the size virtual index. Chunksize defaults to 1MB (1048576 bytes).

##### example	

	curl -X GET [ see Authentication ] http://<host>[:<port>]/node/{id}

##### returns

    {
        "data": {<node>},
        "error": <error message or null>, 
        "status": <http status of request>
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
 - to set file include file field named "upload" containing any file **or** include field named "path" containing the file system path to the file accessible from the Shock server
   
##### example	
  
	curl -X PUT [ see Authentication ] [ -F "attributes=@<path_to_json>" ( -F "upload=@<path_to_data_file>" || -F "path=<path_to_file>") ] http://<host>[:<port>]/node/{id}

  
##### returns

    {
        "data": {<node>},
        "error": <error message or null>, 
        "status": <http status of request>
    }

<br>
**Create index:**

 - Currently available index types include: size (virtual, does not require index creation), line, chunkrecord and record (for sequence file types), and bai (bam index)

##### example	

	curl -X PUT [ see Authentication ] http://<host>[:<port>]/node/{id}/index/<type>
	curl -X PUT [ see Authentication ] http://<host>[:<port>]/node/{id}?index=<type> (deprecated)

##### returns

    {
        "data": null,
        "error": <error message or null>, 
        "status": <http status of request>
    }

##### bam index (bai) argument mapping from URL to samtools

<table border=1>
    <tr>
        <td><b>URL argument</b></td>
        <td><b>value type</b></td>
        <td><b>samtools argument</b></td>
        <td><b>operation</b></td>
    </tr>
    <tr>
        <td>head</td>
        <td>no value</td>
        <td>-h</td>
        <td>Include the header in the output</td>
    </tr>
    <tr>
        <td>headonly</td>
        <td>no value</td>
        <td>-H</td>
        <td>Output the header only.</td>
    </tr>
    <tr>
        <td>count</td>
        <td>no value</td>
        <td>-c</td>
        <td>Instead of printing the alignments, only count them and print the total number.</td>
    </tr>
    <tr>
        <td>flag</td>
        <td>INT</td>
        <td>-f</td>
        <td>Only output alignments with all bits in INT present in the FLAG field.</td>
    </tr>
    <tr>
        <td>lib</td>
        <td>STR</td>
        <td>-l</td>
        <td>Only output reads in library STR</td>
    </tr>
    <tr>
        <td>mapq</td>
        <td>INT</td>
        <td>-q</td>
        <td>Skip alignments with MAPQ smaller than INT</td>
    </tr>
    <tr>
        <td>readgroup</td>
        <td>STR</td>
        <td>-r</td>
        <td>Only output reads in read group STR</td>
    </tr>
</table>

<br>
License
---

Copyright (c) 2010-2012, University of Chicago
All rights reserved.

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

