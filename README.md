About
-----

Shock is a platform to support computation, storage, and distribution. Designed from the ground up to be fast, scalable, fault tolerant, federated. 

Shock is RESTful. Accessible from desktops, HPC systems, exotic hardware, the cloud and your smartphone.

Shock is for scientific data. One of the challenges of large volume scientific data is that without often complex metadata it is of little to no value. Shock allows storage and querying of complex metadata.   

Shock is a data management system. The long term goals of Shock include the ability to annotate, anonymize, convert, filter, perform quality control, and statically subsample at line speed bioinformatics sequence data. Extensible plug-in architecture is in development.

**Most importantly Shock is still very much in development. Be patient and contribute.**

Shock is actively being developed at [github.com/MG-RAST/Shock](https://github.com/MG-RAST/Shock).

Building
--------
Shock (requires mongodb=>2.0.3, go=>1.1.0 [golang.org](http://golang.org/), git, mercurial and bazaar). You must also set the $GOPATH and $GOROOT environment variables before installing Shock. There are two options for installing Shock.

OPTION 1: The recommended method for installing Shock is to download the Makefile located [here](https://raw.github.com/MG-RAST/Shock/master/Makefile) to your $GOPATH directory and then run:

    make install

OPTION 2: You could alternatively install Shock by running:

    go get github.com/MG-RAST/Shock/...

The upside to using OPTION 1 is that this will insert the Shock version number into your Shock server to be displayed when the server is started and this will also generate the Shock documentation locally to be hosted by the server. The built binaries will be located in the env configured $GOPATH/bin/ directory.

### Dockerfile
Alternativly your can use Docker to compile Shock. The Dockerfile in directory "docker" in this repository compiles Shock statically. 
```bash
export TAG=`date +"%Y%m%d.%H%M"`
docker build --force-rm --no-cache --rm -t shock:${TAG} .
```
If you only want the binary you can create an container from the image an copy the binary to your host. This will copy the shock-server binary to your current working directory
```bash
docker run --name shock shock:${TAG} ls
docker cp shock:/go/bin/shock-server .
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

Configuration
-------------
The Shock configuration file is in INI file format. There is a template of the config file located at the root level of the repository.

Running
-------
To run:
  
    shock-server -conf <path_to_config_file>

Documentation
-------------
For further information about Shock's functionality, please refer to our [github wiki](https://github.com/MG-RAST/Shock/wiki/_pages).
