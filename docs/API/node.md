

### API Routes for /node (default port 7445):


#### Nodes

- `/node`
    - `GET` list nodes
- `/node?query`
    - `GET` search query
- `/node/<node_id>`
    - `GET` view node, download file (full or partial)
    - `PUT` modify node (e.g. update attributes of existing node)
    - `DELETE` delete node

#### Node permission

- ` /node/<node_id>/acl`  
    - `GET` view node acls
    - `PUT`  modify node acls
- `/node/<node_id>/acl/<type>`  
    - `GET` view node acls of type `<type>`
    - `PUT` modify node acls of type `<type>`

#### Node indices

- `/node/<node_id>/index/<type>`  
  - `PUT` create node indexes
  - `DELETE` delete node index

- `/node/<node_id>/index/<type>?upload` 
  - `PUT` upload node index
- `/node/<node_id>/acl/<type>?users=<user-ids_or_uuids>`
    - `DELETE`





NOTE: Although a node may be designated as publicly readable, writable, or deletable, user authentication may still be required  to perform the operation depending on the Shock server's configuration.
<br>
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
            "created_on" : "2014-06-16T11:08:17.955-05:00",
            "file": {
                "checksum": {},
                "format": "",
                "name": "",
                "size": 0,
                "virtual": false,
                "virtual_parts": []
            },
            "id": "130cadb5-9435-4bd9-be13-715ec40b2bb5",
            "indexes" : {
                "size" : {
                    "total_units" : 100,
                    "average_unit_size" : 1048576
                }
            }
            "last_modified" : "2014-06-16T11:25:16.535-05:00",
            "linkages" : [],
            "tags" : [],
            "type": basic,
            "version": "aabfee3e4291a649c00984451e1ff891"
        },
        "error": null,
        "status": 200
    }

<br>

### Index:

Currently available index types include: size (virtual, does not require index creation), line, column (for tabbed files), chunkrecord and record (for sequence file types), bai (bam index), and subset (based on an existing index)

##### virtual index:

A virtual index is one that can be generated on the fly without support of precalculated data. The current working example of this 
is the size virtual index. Based on the file size and desired chunksize the partitions become individually addressable. 

##### column index:

A column index is one that is generated on tabbed files which are sorted by the chosen column number.  Each record represents the lines (delimitated by returns) that contain all the same value for the inputted column number.

##### bam index (bai):

To use the bam index feature, the <a href="http://samtools.sourceforge.net/">SAMtools</a> package must be installed on the machine that is running the Shock server with the samtools executable in the path of the user that is running the Shock server.

Also, in order to use this feature, you must sort your .bam file using the 'samtools sort' command before uploading the file into Shock.

##### subset index:

Create a named index based on a list of sorted record numbers that are a subset of an existing index.

<br><br>

API by example
-----------------
All examples use curl but can be easily modified for any http client / library. 
__Note__: Authentication is required for most of these commands
<br>
#### Node Creation ([details](#post_node)):

    # without file or attributes
    curl -X POST http://<host>[:<port>]/node

    # with attributes file
    curl -X POST -F "attributes=@<path_to_json_file>" http://<host>[:<port>]/node

    # with attributes string
    curl -X PUT -F 'attributes_str={ "id": 10 }' http://<host>[:<port>]/node

    # with file, using multipart form
    curl -X POST -F "upload=@<path_to_data_file>" http://<host>[:<port>]/node

    # with file, without using multipart form (not recommended for use with curl!)
    curl -X POST --data-binary @<path_to_data_file> http://<host>[:<port>]/node
        (note: This request format is not recommended for use with curl because curl will read the entire file into memory before sending it. Conversely, other programming languages and applications have the opposite issue, reading the entire file into memory for a form POST but not this POST format.)
        (also note: Posting an empty file in this way will result in an empty node with no file rather than an empty node with an empty file)

    # setting location tag
    curl -X POST -H 'Authorization: <secret>'  -H "Content-Type: application/json"  http://<host>[:<port>]/node/${id}/locations -d '{"id":"Location1" }'
    (note: this sets a location id Location1 for the node ${id}; Locations are defined in the locations.yaml config file)

    # example with optional requested boolean and requestedDate for node 96576d58-6e2d-4bf5-8edf-8224cf29291c
    curl -X POST -H 'Authorization: <secret>'  -H "Content-Type: application/json"  "localhost:7445/node/96576d58-6e2d-4bf5-8edf-8224cf29291c/locations/" -d '{"id":"test1" ,  "stored": true, "requestedDate": "2018-09-22T12:42:31+07:00" }' ,  "requested": true, "requestedDate": "2018-09-22T12:42:31+07:00" }'

    # with gzip compressed file, to be uncompressed in node
    curl -X POST -F "gzip=@<path_to_data_file>" http://<host>[:<port>]/node

    # with bzip2 compressed file, to be uncompressed in node
    curl -X POST -F "bzip2=@<path_to_data_file>" http://<host>[:<port>]/node

    # create node by copying data file from another node, optionally specify copy_indexes=1 to additionally copy indexes from parent node
    curl -X POST -F "copy_data=<copy_node_id>" http://<host>[:<port>]/node

    # create a "subset" node which is a node where the data source is composed of a subset of indices from a parent node
    curl -X POST -F "parent_node=<parent_node_id>" -F "parent_index=<index>" -F "subset_indices=@<path_to_file>" http://<host>[:<port>]/node

    # copying node from one shock server to another shock server, by default this copies data and attributes - add &post_data=0 and/or &post_attr=0 to the url to disable either
	(note: if the destination shock server requires authentication, you must provide authentication in your GET request and the credentials will be passed along to the destination shock server.)
    curl -X GET http://<host>[:<port>]/node/<node_id>?download_post&post_url=http://<destination_host>[:<destination_port>]/node

    # with file local to the shock server
    curl -X POST -F "path=<path_to_data_file>" -F "action=<action_type>" http://<host>[:<port>]/node
        (note: The action_type is one of keep_file (node points to file path given), copy_file (file is copied to shock data directory), or move_file (file is moved to shock data directory).  The move_file action only works if user running Shock has permissions to move the file.)
    
    # with file upload in N parts (part uploads may be done in parallel and out of order)
    curl -X POST -F "parts=N" -F "file_name=<file_name>" http://<host>[:<port>]/node
    curl -X PUT -F "1=@<file_part_1>" http://<host>[:<port>]/node/<node_id>
    curl -X PUT -F "2=@<file_part_2>" http://<host>[:<port>]/node/<node_id>
    ...
    curl -X PUT -F "N=@<file_part_N>" http://<host>[:<port>]/node/<node_id>
	
    # with file upload in N parts where N is unknown at node creation time (part uploads may be done in parallel and out of order)
    curl -X POST -F "parts=unknown" -F "file_name=<file_name>" http://<host>[:<port>]/node
    curl -X PUT -F "1=@<file_part_1>" http://<host>[:<port>]/node/<node_id>
    curl -X PUT -F "2=@<file_part_2>" http://<host>[:<port>]/node/<node_id>
    ...
    curl -X PUT -F "parts=close" http://<host>[:<port>]/node/<node_id>
        (note: file_name is an optional parameter for files uploaded in parts. The file name will default to the node id if it is not set.)

    # with compressed file in N parts (unknown or given), to be uncompressed in node when parts completed
    curl -X POST -F "parts=N" -F "compression=gzip" http://<host>[:<port>]/node
    curl -X POST -F "parts=unknown" -F "compression=bzip2" http://<host>[:<port>]/node

    # create multiple nodes from a node with an archive file, supports: zip, tar, tar.gz, tar.bz2
    # if an attributes file is included it will be applied to all the created nodes
    curl -X POST -F "unpack_node=<archive_node_id>" -F "archive_format=<format>" http://<host>[:<port>]/node

<br>

#### Node retrieval ([details](#get_node)):

    # node information
    curl -X GET http://<host>[:<port>]/node/<node_id>

    # download file
    curl -X GET http://<host>[:<port>]/node/<node_id>?download

    # download file as data stream (not as a file attachment)
    curl -X GET http://<host>[:<port>]/node/<node_id>?download_raw

    # download first 1mb of file
    curl -X GET http://<host>[:<port>]/node/<node_id>?download&index=size&part=1
        
    # download first 10mb of file
    curl -X GET http://<host>[:<port>]/node/<node_id>?download&index=size&chunk_size=10485760&part=1

    # download Nth 10mb of file
    curl -X GET http://<host>[:<port>]/node/<node_id>?download&index=size&chunk_size=10485760&part=N
	
    # download portion of file given seek and length positions (in bytes)
    curl -X GET http://<host>[:<port>]/node/<node_id>?download&seek=<seek>&length=<length>
        (note: exluding seek position defaults to an offset of zero bytes, exluding length position defaults to remainder of file being returned)

    # download fastq sequence file in fasta format
    curl -X GET http://<host>[:<port>]/node/<node_id>?download&filter=fq2fa

    # download sequence file (fasta or fastq) with anonymous unique header IDs
    curl -X GET http://<host>[:<port>]/node/<node_id>?download&filter=anonymize

    # download file in compressed format, works with all the above options
    curl -X GET http://<host>[:<port>]/node/<node_id>?download&compression=<zip|gzip>

    # retrieve pre-authorized download url (returns 1-time use download url that does not require auth and is valid for 24 hours)
    curl -X GET http://<host>[:<port>]/node/<node_id>?download_url

    # download multiple files in a single archive format (zip or tar), returns 1-time use download url for archive
    # use download_url with a standard query
    curl -X GET http://<host>[:<port>]/node?query&download_url&archive=zip&<key>=<value>
    # use download_url with a POST and list of node ids
    curl -X POST -F "download_url=1" -F "archive_format=zip" -F "ids=<node_id_1>,<node_id_2>,<...>" http://<host>[:<port>]/node

    # download entire bam file in human readable sam alignments
    curl -X GET http://<host>[:<port>]/node/<node_id>?download&index=bai

    # download bam alignments overlapped with specified region (ref_id:start_pos-end_pos)
    curl -X GET http://<host>[:<port>]/node/<node_id>?download&index=bai&region=chr1:1-20000

    # download bam alignments with selected arguments supported by "samtools view"
    curl -X GET http://<host>[:<port>]/node/<node_id>?download&index=bai&head&headonly&count&flag=[INT]&lib=[STR]&mapq=[INT]&readgroup=[STR]
        (note: All the arguments are optional and can be used with or without the region, but the index=bai is required)
    
<br>

#### Node acls: 

    # view all acls
    curl -X GET http://<host>[:<port>]/node/<node_id>/acl/

    # view specific acls
    curl -X GET http://<host>[:<port>]/node/<node_id>/acl/[ all | read | write | delete | owner ]

    # changing owner (chown)
    curl -X PUT http://<host>[:<port>]/node/<node_id>/acl/owner?users=<user-id_or_uuid>

    # adding user to acls
    curl -X PUT http://<host>[:<port>]/node/<node_id>/acl/[ all | read | write | delete ]?users=<user-ids_or_uuids>  

    # deleting user to acls
    curl -X DELETE http://<host>[:<port>]/node/<node_id>/acl/[ all | read | write | delete ]?users=<user-ids_or_uuids>

<br>

#### Querying ([details](#get_nodes)):

    # by attribute key value
    curl -X GET http://<host>[:<port>]/node?query&<key>=<value>

    # by attribute key value, limit 10
    curl -X GET http://<host>[:<port>]/node?query&<key>=<value>&limit=10

    # by attribute key value, limit 10, offset 10
    curl -X GET http://<host>[:<port>]/node?query&<key>=<value>&limit=10&offset=10

    # by any key value (this allows querying of fields outside of attributes section)
    curl -X GET http://<host>[:<port>]/node?querynode&<key>=<value>

    # by ACL's (enter users-ids or uuids as comma-separated list, this works for query or querynode)
    curl -X GET http://<host>[:<port>]/node?querynode&[ owner | read | write | delete ]=<user-ids_or_uuids>
        (note: resultant set is a subset of the nodes that are viewable to the authenticated user)

    # by public ACL's (returns nodes that have a public setting for the given ACL)
    curl -X GET http://<host>[:<port>]/node?querynode&[ public_owner | public_read | public_write | public_delete ]=1

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

<br>

### GET /

Description of resources available through this api

##### example
	
    curl -X GET http://<host>[:<port>]/
	
##### returns

    {"resources":["node"],"url":"http://localhost:7445/","documentation":"http://localhost:7445/documentation.html","contact":"admin@host.com","id":"Shock","type":"Shock"}

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

<br>

### GET /node/<node_id>

View node, download file (full or partial)

 - optionally takes user/password via Basic Auth
 - ?download - complete file download
 - ?download&index=size&part=1\[&part=2...\]\[chunksize=inbytes\] - download portion of the file via the size virtual index. Chunksize defaults to 1MB (1048576 bytes).

##### example	

	curl -X GET [ see Authentication ] http://<host>[:<port>]/node/<node_id>

##### returns

    {
        "data": {<node>},
        "error": <error message or null>, 
        "status": <http status of request>
    }

<br>

### PUT /node/<node_id>

Modify node, create index

 - optionally takes user/password via Basic Auth
 
**Modify:** 

 - **Once the file of a node is set, it is immutable.**
 - node attributes can be over-written
 - accepts multipart/form-data encoded 
 - to set attributes include file field named "attributes" containing a json file of attributes
 - to set file include file field named "upload" containing any file **or** include field named "path" containing the file system path to the file accessible from the Shock server
   
##### example	
  
	# update attributes
	curl -X PUT -F "attributes=@<path_to_json>" http://<host>[:<port>]/node/<node_id>
	
	# add file
	curl -X PUT ( -F "upload=@<path_to_data_file>" || -F "path=<path_to_file>") http://<host>[:<port>]/node/<node_id>
        
	# change filename
	curl -X PUT -F "file_name=<new_file_name>" http://<host>[:<port>]/node/<node_id>
	
	# add / update expiration
	curl -X PUT -F "expiration=<\d+[MHD]>" http://<host>[:<port>]/node/<node_id>

	# remove expiration
	curl -X PUT -F "remove_expiration=true" http://<host>[:<port>]/node/<node_id>

  
##### returns

    {
        "data": {<node>},
        "error": <error message or null>, 
        "status": <http status of request>
    }

<br>

**Create index:**

 - Currently available index types include: size (virtual, does not require index creation), line, column (for tabbed files), chunkrecord and record (for sequence file types), bai (bam index), and subset (based on an existing index)

##### example	
    
	curl -X PUT [ see Authentication ] http://<host>[:<port>]/node/<node_id>/index/<type>
	curl -X PUT [ see Authentication ] http://<host>[:<port>]/node/<node_id>/index/column?number=<int>
	curl -X PUT [ see Authentication ] -F "index_name=<string>" -F "parent_index=<type>" -F "subset_indices=@<path_to_file>" http://<host>[:<port>]/node/<node_id>/index/subset
	curl -X PUT [ see Authentication ] http://<host>[:<port>]/node/<node_id>?index=<type> (deprecated)

	If an index already exists, you should receive an error message telling you that.  To overwrite the existing index, add the parameter '?force_rebuild=1' to your PUT request.

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

**Upload index:**

To upload an index, query field `?upload` and form field `upload=` both have to be used.

    
    curl -X PUT [ see Authentication ] -F "upload=@<path_to_index_file>" http://<host>[:<port>]/node/<node_id>/index/<type>?upload
    
    curl -X PUT [ see Authentication ] -F "upload=@<path_to_index_file>" http://<host>[:<port>]/node/<node_id>/index/<type>?upload&indexFormat=<string>&avgUnitSize=<int64>&totalUnits=<int64>
    





