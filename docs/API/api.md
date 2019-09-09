
## Shock API 

The Shock API provides the following resources

- [/node ](./node.md)
- [/location](./location.md)
- [/types](./types.md)

Info on Authentication and Authentication is [here](./Authorization.md). 

Follow the links above for more details.

## Basic Examples for interacting with Shock
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