
# Shock 2.0

An object store for scientific data that is:

-  a storage platform designed from the ground up to be fast, scalable and fault tolerant.

-  fully RESTful. The [API](./API/README.md) is accessible from desktops, HPC systems, exotic hardware, the cloud and your smartphone.

-  designed from scratch for complex scientific data and allows the storage and querying of complex user-defined metadata.   

- a full data management system that supports in storage layer operations like quality-control, format conversion, filtering or subsetting.

- integrated with S3, Microsoft Azure Storage, Google Cloud Storage and IBM's Tivoli TSM storage managment system.

- integrated with HSM operations and caching

-  part of our reproducible science platform [Skyport]([https://github.com/MG-RAST/Skyport2) combined to create [Researchobjects](http://www.researchobject.org/) when combined with [CWL](http://www.commonwl.org) and 
[CWLprov](https://github.com/common-workflow-language/cwlprov)

Shock is actively being developed at [github.com/MG-RAST/Shock](https://github.com/MG-RAST/Shock).

(see [Shock: Active Storage for Multicloud Streaming Data Analysis](http://ieeexplore.ieee.org/abstract/document/7406331/), Big Data Computing (BDC), 2015 IEEE/ACM 2nd International Symposium on, 2015)

Check out the notes on [building and installing Shock](./building.md) and [configuration](./configuration.md).


## Shock in 30 seconds (for Docker-compose) (or later for kubectl )
We know that you already have `docker` and `docker-compose` installed and `curl` is available locally. To improve readability of JSON output we recommend using a JSON pretty printer such as [`jq`](https://stedolan.github.io/jq/download/) or `python -m json.tool` 

### Start Shock
```bash
docker-compose up
```
This will automatically download and start Shock and MongoDB Dockerimages. Note that in this demo-configuration Shock does not store data persistently.

### Test connection to Shock

Open another terminal: 

```bash
curl http://localhost:7445/ | jq .
```
Should return a JSON object describing Shock, i.e. version number.

### Push a file into Shock

```bash
curl -X POST -F 'upload=@test/testdata/10kb.fna' http://localhost:7445/node | jq .
```

returns

```text
{
  "status": 200,
  "data": {
    "id": "8eb28ad3-2561-4847-8034-6d5473fecfad",
    "version": "88e3227e7ba47a8d595f243ef16f9c2b",
    "file": {
      "name": "10kb.fna",
      "size": 11914,
      "checksum": {
        "md5": "730c276ea1510e2b7ef6b682094dd889"
      },
      "format": "",
      "virtual": false,
      "virtual_parts": null,
      "created_on": "2019-09-23T20:22:52.506403Z",
      "locked": null
    },
    "attributes": null,
    "indexes": {
      "size": {
        "total_units": 1,
        "average_unit_size": 1048576,
        "created_on": "2019-09-23T20:22:52.5165769Z",
        "locked": null
      }
    },
    "version_parts": {
      "acl_ver": "b46701cc24139e5cca2100e09ec48c19",
      "attributes_ver": "2d7c3414972b950f3d6fa91b32e7920f",
      "file_ver": "fcd9613da51b9b181ff94434a19add87",
      "indexes_ver": "88455c093e82651aa042252dca2a37f8"
    },
    "tags": null,
    "linkage": null,
    "priority": 0,
    "created_on": "2019-09-23T20:22:52.5193448Z",
    "last_modified": "0001-01-01T00:00:00Z",
    "expiration": "0001-01-01T00:00:00Z",
    "type": "basic",
    "parts": null,
    "locations": null
  },
  "error": null
}
```

The resulting JSON object contains an `id` field (line 4 in this example), in this example its value is `8eb28ad3-2561-4847-8034-6d5473fecfad`. This identifier is a [uuid](https://en.wikipedia.org/wiki/Universally_unique_identifier), which can be used to download the file. 

Saving the node id in an environment variable allows to simply copy-and-paste the following examples:

```bash
NODE_ID=<uuid>
```


### View Shock node


```bash
curl http://localhost:7445/node/${NODE_ID} | jq .
```

### Download a file


```bash
curl -OJ "http://localhost:7445/node/${NODE_ID}?download"
```

returns: `curl: Saved to filename '10kb.fna'`

Option `-OJ` makes sure that curl saves the file using the correct filename.


### Add metadata

```bash
curl -X PUT -F 'attributes_str={"project":"extraterrestrial_lifeforms", "sample-nr": 1}' http://localhost:7445/node/${NODE_ID} | jq .
```

### Search by metadata

List all nodes in the project:

```bash
curl 'http://localhost:7445/node?query&project=extraterrestrial_lifeforms' | jq .
```


## Documentation
- [API documentation](./API/README.md).
- [Building shock](./building.md).
- [Configuring](./configuration.md).
- [Concepts](./concepts.md).
- [Caching and data migration](./caching_and_data_migration.md).
- For further information about Shock's functionality, please refer to our [Shock documentation](https://github.com/MG-RAST/Shock/docs/).

