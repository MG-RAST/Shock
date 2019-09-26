## Shock API specification

The Shock API provides the following resources

- [/node](api.html#api-Node)
- [/location](./location.md)
- [/types](./types.md)



[Shock API specification](api.html)



Info on Authentication and Authentication is [here](./Authorization.md). 

Follow the links above for more details.

## Basic Examples for interacting with Shock



[Examples for node creation / file upload](api.html#api-Node-nodePost)


[Examples for node retrival / file download](api.html#api-Node-api-Node-nodeNodeIdGet)


[Examples for node search](api.html#api-Node-nodeGET)
    

#### Node acls: 

[Examples view permissions](api.html#api-Node-nodeNodeidAclGet)

[Examples to set permissions](api.html#api-Node-nodeNodeidAclPUT)


#### Node incides: 

[Examples to create indices](api.html#api-Node-nodeNodeidIndexTypePut)



## API repsonse


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





## TODO move into api spec


Files in Shock are stored as Nodes with a file and a metadata set. An example metadata set is shown below:

A command like the following:
`curl -X GET http://<host>[:<port>]/node/130cadb5-9435-4bd9-be13-715ec40b2bb5`
will download a JSON structure like below.
~~~~
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
~~~~
Metadata is free form, however certain fields will be indexed by the MongoDB and therefore will make retrieval very efficient. The Shock server will report the names of index keys when invoke without parameters,


### Upload a file
    # the most simple case, upload `myfile` to `myserver`
    curl -X PUT -F 'file_name=@myfile' http://<myserver>[:<port>]/node

    # add metadata to myfile
    curl -X POST  -F 'attributes_str={ "mymetadata_field": "myvalue" }' -F "file_name=@myfile" http://<host>[:<port>]/node

For both cases it might a good idea to check the MD5 checksums that Shock computes after upload.

### Download a file
A file is stored as a node with a unique ID in Shock.

    # download the file for <node_id>
    curl --silent -X GET -H "${AUTH}" "${SHOCK_SERVER}/node/<node_id/?download" 
    
It is good practice to ask for the server side MD5 checksum and compare it to a local checksum.

### Add metadata to a file stored in Shock
Add metadata in JSON format for an existing node.

    # with attributes file
    curl -X POST -F "attributes=@<path_to_json_file>" http://<host>[:<port>]/node/<node_id>

or a little more convenient in many cases:

    # with attributes string
    curl -X PUT -F 'attributes_str={ "id": 10 }' http://<host>[:<port>]/node/<node_id>

### Search for files
Obtain a list of nodes matching a query.

    # obtain a list of nodes where <key> has <value>
    curl -X GET http://<host>[:<port>]/node?query&<key>=<value>

or return a limited number of entries

    # same as  above but only return 10 items
    curl -X GET http://<host>[:<port>]/node?query&<key>=<value>&limit=10

Please NOTE: The links at the top of the page provide A LOT more detail.





### Available index types

Currently available index types include: size (virtual, does not require index creation), line, column (for tabbed files), chunkrecord and record (for sequence file types), bai (bam index), and subset (based on an existing index)

##### virtual index

A virtual index is one that can be generated on the fly without support of precalculated data. The current working example of this 
is the size virtual index. Based on the file size and desired chunksize the partitions become individually addressable. 

##### column index

A column index is one that is generated on tabbed files which are sorted by the chosen column number.  Each record represents the lines (delimitated by returns) that contain all the same value for the inputted column number.

##### bam index (bai)

To use the bam index feature, the <a href="http://samtools.sourceforge.net/">SAMtools</a> package must be installed on the machine that is running the Shock server with the samtools executable in the path of the user that is running the Shock server.

Also, in order to use this feature, you must sort your .bam file using the 'samtools sort' command before uploading the file into Shock.

##### subset index

Create a named index based on a list of sorted record numbers that are a subset of an existing index.

##### bam index (bai) argument mapping from URL to samtools

Table 

<table>
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



  

