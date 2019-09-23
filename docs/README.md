
# Shock 2.0 -- An object store for scientific data

-  is a storage platform designed from the ground up to be fast, scalable and fault tolerant.

- is RESTful. The [API](./API/README.md) is accessible from desktops, HPC systems, exotic hardware, the cloud and your smartphone.

- is designed for complex scientific data and allows storage and querying of complex user-defined metadata.   

- is a data management system that supports in storage layer operations like quality-control, format conversion, filtering or subsetting.

- is integrated with S3, Microsoft Azure Storage, Google Cloud Storage and IBM's Tivoli TSM storage managment system.

- supports HSM operations and caching

- is part of our reproducible science platform [Skyport]([https://github.com/MG-RAST/Skyport2) combined to create [Researchobjects](http://www.researchobject.org/) when combined with [CWL](http://www.commonwl.org) and 
[CWLprov](https://github.com/common-workflow-language/cwlprov)

Shock is actively being developed at [github.com/MG-RAST/Shock](https://github.com/MG-RAST/Shock).

(see [Shock: Active Storage for Multicloud Streaming Data Analysis](http://ieeexplore.ieee.org/abstract/document/7406331/), Big Data Computing (BDC), 2015 IEEE/ACM 2nd International Symposium on, 2015)

Check out the notes  on [building and installing Shock](./building.md) and [configuration](./configuration.md).


## Shock in 30 seconds (for Docker-compose) (or later for kubectl )
This assumes that you have `docker` and `docker-compose` installed and `curl` is available locally.

### Download the container
`docker-compose up`

Don't forget to later `docker-compose down` and do not forget, by default this configuration does not store data persistently.

### Push a file into Shock
`curl -H 'Authorization: basic dXNlcjE6c2VjcmV0' -X PUT -F 'file_name=@myfile' http://localhost:7445/node`


### Download a file from Shock

`curl -H 'Authorization: basic dXNlcjE6c2VjcmV0' http://localhost:7445/node`

Do not forget, by default this configuration does not store data persistently.

## Documentation
- [API documentation](./API/README.md).
- [Building shock](./building.md).
- [Configuring](./configuration.md).
- [Concepts](./concepts.md).
- [Caching and data migration](./caching_and_data_migration.md).
- For further information about Shock's functionality, please refer to our [Shock documentation](https://github.com/MG-RAST/Shock/docs/).

