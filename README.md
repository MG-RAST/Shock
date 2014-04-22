About:
------

Shock is a platform to support computation, storage, and distribution. Designed from the ground up to be fast, scalable, fault tolerant, federated. 

Shock is RESTful. Accessible from desktops, HPC systems, exotic hardware, the cloud and your smartphone.

Shock is for scientific data. One of the challenges of large volume scientific data is that without often complex metadata it is of little to no value. Shock allows storage and querying of complex metadata.   

Shock is a data management system. The long term goals of Shock include the ability to annotate, anonymize, convert, filter, perform quality control, and statically subsample at line speed bioinformatics sequence data. Extensible plug-in architecture is in development.

**Most importantly Shock is still very much in development. Be patient and contribute.**

Shock is actively being developed at [github.com/MG-RAST/Shock](https://github.com/MG-RAST/Shock).

Building:
---------
Shock (requires mongodb=>2.0.3, go=>1.1.0 [golang.org](http://golang.org/), git, mercurial and bazaar):

    go get github.com/MG-RAST/Shock/...

Built binary will be located in env configured $GOPATH or $GOROOT depending on the Go configuration.

Configuration:
--------------
The Shock configuration file is in INI file format. There is a template of the config file located at the root level of the repository.

Running:
--------------
To run:
  
    shock-server -conf <path_to_config_file>

More:
--------------
For further information about Shock's functionality, please refer to our [github wiki](https://github.com/MG-RAST/Shock/wiki/_pages).
