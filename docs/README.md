
# Shock -- a storage manager for data science

-  is a storage platform designed from the ground up to be fast, scalable and fault tolerant.

- is RESTful. Accessible from desktops, HPC systems, exotic hardware, the cloud and your smartphone.

- is designed for complex scientific data and allows storage and querying of complex user-defined metadata.   

- is a data management system that supports in storage layer operations like quality-control, format conversion, filtering or subsetting.

- is integrated with S3, IBM's Tivoli TSM storage managment system.

- supports HSM operations and caching

Shock is actively being developed at [github.com/MG-RAST/Shock](https://github.com/MG-RAST/Shock).

(see [Shock: Active Storage for Multicloud Streaming Data Analysis](http://ieeexplore.ieee.org/abstract/document/7406331/), Big Data Computing (BDC), 2015 IEEE/ACM 2nd International Symposium on, 2015)

Check out the notes  on [building and installing Shock](./building.md) and [configuration](./configuration.md).


## Shock in 30 seconds (for Docker-compose)
This assumes that you have `docker` and `docker-compose` installed and `curl` is available locally.

### Download the container
`docker-compose up`

Don't forget to later `docker-compose down` and do not forget, by default this configuration does not store data persistently.

### Push a file into Shock
`curl -H 'Authorization: basic dXNlcjE6c2VjcmV0' -X PUT -F 'file_name=@myfile' http://localhost:7445/node`


### Download a file from Shock

`curl -H 'Authorization: basic dXNlcjE6c2VjcmV0' http://localhost:7445/node`


## Shock in 30 seconds (for Kubernetes)
This assumes that you have `docker` and `docker-compose` installed and `curl` is available locally.

### setup the namespace and create resources
`kubectl create shock-setup.yaml`
### set up the shock service 
`kubectl create shock-service.yaml`
### check
`kubectl get -n shock get pods

...

Do not forget, by default this configuration does not store data persistently.

### Push a file into Shock
`curl -H 'Authorization: basic dXNlcjE6c2VjcmV0' -X PUT -F 'file_name=@myfile' http://<URL>/node`


### Download a file from Shock

`curl -H 'Authorization: basic dXNlcjE6c2VjcmV0' http://<URL>node`


## Documentation
- [API documentation](./API/README.md).
- [Building shock](./building.md).
- [Configuring Shock](./configuration.md).
- [Shock clients](./Shock-Clients.md)
- [Shock concepts](./Concepts.md).
- [Caching and data migration](./caching_and_data_migration.md).
- For further information about Shock's functionality, please refer to our [Shock documentation](https://github.com/MG-RAST/Shock/docs/).

