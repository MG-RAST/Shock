About
-----

Shock is a platform to support computation, storage, and distribution. Designed from the ground up to be fast, scalable, fault tolerant, federated.  (see [Shock: Active Storage for Multicloud Streaming Data Analysis](http://ieeexplore.ieee.org/abstract/document/7406331/), Big Data Computing (BDC), 2015 IEEE/ACM 2nd International Symposium on, 2015)

Shock is RESTful. Accessible from desktops, HPC systems, exotic hardware, the cloud and your smartphone.

Shock is for scientific data. One of the challenges of large volume scientific data is that without often complex metadata it is of little to no value. Shock allows storage and querying of complex metadata.   

Shock is a data management system. The long term goals of Shock include the ability to annotate, anonymize, convert, filter, perform quality control, and statically subsample at line speed bioinformatics sequence data. Extensible plug-in architecture is in development.


Shock is actively being developed at [github.com/MG-RAST/Shock](https://github.com/MG-RAST/Shock).

Building
--------
To build Shock manually, use the Makefile. Note that you need golang (>=1.6.0). 

### Docker

You can get the Shock Dockerimage with:
```bash
docker pull mgrast/shock
```

Or, to build the Docker image on your own:
```bash
git clone --recursive https://github.com/MG-RAST/Shock.git
cd Shock
docker build --force-rm --no-cache --rm -t mgrast/shock .
```

If you only need the statically compiled binary, you can extract it from the Dockerimage:
```bash
VERSION=$(docker run --rm mgrast/shock shock-server --version | grep version | grep -o v[0-9].* |tr -d '\n')
echo $VERSION
docker create --name shock mgrast/shock
mkdir -p bin
docker cp shock:/go/bin/shock-server ./bin/shock-server-${VERSION}
docker rm shock
```

### MongoDB

In ubuntu you can simply install mongo with:
```bash
sudo apt-get install -y mongodb-server
```
If you do not want to use a package manager to install mongodb, use:
```bash
curl -s http://downloads.mongodb.org/linux/mongodb-linux-x86_64-2.4.14.tgz | tar -v -C /mongodb/ -xz
```
If you do not use a service manager such as systemd, you can start mongodb like this, in foreground:
```bash
/mongodb/bin/mongod --dbpath /data/
```
or in background:
```bash
nohup /mongodb/bin/mongod --dbpath /mnt/db/ &
```
You can also run MongoDB in a docker container:
```bash
mkdir -p /mnt/shock-server/mongodb
export DATADIR="/mnt/shock-server"
docker run --rm --name shock-server-mongodb -v ${DATADIR}/mongodb:/data/db --expose=27017 mongo mongod --dbpath /data/db
```


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

Documentation
-------------
For further information about Shock's functionality, please refer to our [github](https://github.com/MG-RAST/Shock/).

Developer Notes
---------------

To update vendor directory use the tool govendor: `go get -u github.com/kardianos/govendor`
