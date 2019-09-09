

Configuration
-------------
The Shock configuration file is in INI file format. There is a template of the config file located at the root level of the repository.

Running
-------
To run:
```bash
shock-server -conf <path_to_config_file>
```
With docker:
```bash
mkdir -p /mnt/shock-server/log
mkdir -p /mnt/shock-server/data
export DATADIR="/mnt/shock-server"
docker run --rm --name shock-server -p 7445:7445 -v ${DATADIR}/shock-server.cfg:/shock-config/shock-server.cfg -v ${DATADIR}/log:/var/log/shock -v ${DATADIR}/data:/usr/local/shock --link=shock-server-mongodb:mongodb mgrast/shock /go/bin/shock-server --conf /shock-config/shock-server.cfg
```
Comments:<br>
port 7445: Shock server API (default in config)<br>
"-v" mounts host to container directories<br>
"--link" connects Shock server and mongodb (--link=$imagename:$alias) so you need to put the alias (in the example "mongodb") as the value of the hosts variable in the shock-server.cfg
The parameters in the following determine the data migration and caching properties of the Shock server
- Locations.yaml file

- PATH_CACHE parameter

If the `PATH_CACHE` parameter parameter is enabled, Shock will attempt to download nodes present in Mongo that are NOT present on local disk in `PATH_DATA` 
from one of the Locations (see [Concepts]./Concepts.md) configured. The node will point to the correct location with the node data. Both index and file are 
restored to the local system and stored in the `PATH_CACHE` directory.


- Types.yaml file
