Shock 
=====

To build:
---------

Unix/Macintosh 

To install go weekly.2012-01-20 (http://weekly.golang.org/doc/install/source):
    
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
  
    ./bin/shock-server -port=<port#> -dataroot=<path_to_data_root>
  
API
---

### Node:

#### id
unique identifier

#### file_name 
file name for attached file if present

#### size
file size for attached file if present

#### checksum
file checksum for attached file if present

#### attributes
arbitrary json

#### acl
access control (in development)

#### node example:

	{
	    "id": 6775,
	    "file_name": "h_sapiens_asm.tar.gz",
	    "checksum": "8fd07ad670159c491eed7baefe97c16a",
	    "size": 2819582549,
	    "attributes": {
	        "description": "tar gzip of h_sapiens_asm bowtie indexes",
	        "file_list": [
	            "h_sapiens_asm.1.ebwt",
	            "h_sapiens_asm.2.ebwt",
	            "h_sapiens_asm.3.ebwt",
	            "h_sapiens_asm.4.ebwt",
	            "h_sapiens_asm.rev.1.ebwt",
	            "h_sapiens_asm.rev.2.ebwt"
	        ],
	        "source": "ftp://ftp.cbcb.umd.edu/pub/data/bowtie_indexes/h_sapiens_asm.ebwt.zip"
	    },
	    "acl": {
	        "read": [],
	        "write": [],
	        "delete": []
	    }
	}

### Actions
### Create node:
POST /node (multipart/form-data encoded)

 - to set attributes include file field named "attributes" containing a json file of attributes
 - to set file include file field named "file" containing any file

#### example
	
	curl -X POST [ -F "attributes=@<path_to_json>" -F "file=@<path_to_data_file>" ] <shock_url>[:<port>]/node
	
#### returns

	<new_node>

<br/>
### List nodes:
GET /node

 - by adding ?offset=N you get the nodes starting at N+1 (in development)
 - by adding ?count=N you get a maximum of N nodes returned (in development)

#### example
	
	curl -X GET <shock_url>[:<port>]/node/[?offset=<offset>&count=<count>]
		
#### returns

	{"total_nodes":42,"offset":0,"count":4,"nodes":[<node_1>, <node_2>, <node_3>, <node_4>]}

<br/>	
### Get node:
GET /node/<nodeid>
	
 - ?download - complete file download
 - ?download&index=$index&part=$part - file part download (in development)
 - ?list&indexes - list available indexes (in development)
 - ?list&index=$index - index parts list	(in development)
	
#### example	

	curl -X GET <shock_url>[:<port>]/node/<nodeid>
	
#### returns

	<node>

