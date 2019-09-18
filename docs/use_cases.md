# Use cases for Shock

Shock was created as a hybrid between a traditional file store and a database. It offers the best of both worlds. All files can be stored immediately with almost zero overhead. 
But if desired, etadata can be added allowing for a head start on the required business logic.

## Shock Origin story

### Indexing and Expiration
We designed Shock initially to work with our MG-RAST workflow. At that time implemented as a series of shell scripts and scheduled by Sun Grid engine or Slurm or something (I honestly do not remember).

One of our most frequent problems was: I need to delete all temporary files for jobs X .... Y. 

Two problems arise: Where are those jobs and which files are temporary. We note that our pipeline is and was constantly evolving, simply use a regular expression was (and is) not an option.

Shock implements a FileReaper() function that expires nodes when their TTL is over. 

### Firewalls and a bit of history
Shock was initially created around the year 2009 and one of the main driving forces was our frustation with installing and maintaining large shared file systems. We operated a large NFS store, a GPFS store, A Lustre store. Each of them had good sides, but every one of them was not designed for the scale of our opertion (many hundreds of gigabytes at the time). In addition, we are experts at stealing computational cycles wherever they can be found (sound familiar in cloud times?), with traditional shared file systems the following dialogue got really old: "Ok, we need to mount X on Y and need to have a number of ports openend up on the firewall(s)." Response: "No we cannot do that, why don't you ...".

Shock is our answer to this dilemma. All you need is port 80. Ah and btw. proxies work as does chaching. All we use is HTTP.


## Some example use cases

Examples are the workflows in MG-RAST, for each metagenome we run a multi-stage (as in ~20) step analytical pipeline. Each steps leaves a data trace of small and large files. We can store all the files, add a TTL (time to live, effectively a file expiration data) and never worry about them. We use this to keep intermediate files around for a few days in case something went wrong. 


## Example 1: MG-RAST
### requirements:
- ~ 1PB of files
- indexing
- automated deletion of intermediary results (expiry) after set intervals
- running on commodity hardware
- support line speed or near line speed

### Our implementation:

- set up a dedicated Server with initially 1PB of local storage on commodity hardware
- the local storage is on a ZFS filesystem (ZFS on Linux)
- added an S3 backend
- added a TSM backend store
- set up node removal from the main store IF a node is stored in two locations (S3 and TSM)
- we achieve 6-7Gbit/s without optimization

We created the following types:
~~~~
- ID: "EBI Submission Receipt"
  Description: "EBI Submission Receipt"
  Priority: 7
- ID: "inbox"
  Description: "MG-RAST inbox node"
  Priority: 1
- ID: "metagenome"
  Description: "MG-RAST metagenome"
  Priority: 9
  Data-Types:
    - fa
    - fasta
    - fastq
    - fq
    - bam
    - sam
- ID: "submission"
  Description: "MG-RAST submission"
  Priority: 9
- ID: "cv"
  Description: "Controlled Vocabulary"
  Priority: 7
- ID: "ontology"
  Description: "ontology"
  Priority: 7
- ID: "backup"
  Description: "Backup or Dump from another system e.g. MongoDB or MySQL"
  Priority: 9
- ID: "metadata"
  Description: "metadata"
  Priority: 7
- ID: "mixs"
  Description: "GSC MIxS Metadata file XLS format"
  Priority: 9
  Data-Types:
    - xls
    - xlsx
    - json
- ID: "reference"
  Description: "reference database"
  Priority: 7
- ID: "analysisObject"
  Description: "MG-RAST analysisObject"
  Priority: 1
- ID: "analysisRecipe"
  Description: "MG-RAST analysisRecipe"
  Priority: 1
- ID: "preference"
  Description: "MG-RAST user preference"
  Priority: 1
~~~~
## Example 2: DNA sequencing Service

To create a low maintenance, low cost store for a DNA sequencing service facility we chose Shock as the underlying data store.

The requirements included:
- use of commodity hardware
- cost effectiveness
- simplicity of use
- integration with current workflows and analysis software
- integration with vendor provided instrument integration (e.g. Windows shares)

### Our implementation:

- set up a dedicated Shock server with an S3 Backend and Tape backup via IBM TSM
- creation of a series of shell scripts for data import and export into shock
- we also created CWL workflows for processing

- introduced three new "types" in Shock
 - raw sequence data with no expiration
 - intermediate data (image thumbnails useful for debugging) with expiration time of 2 weeks
 - settings file for each DNA sequencing instrument run

We implemented the following data types
 ~~~~
 - ID: "run-folder-archive-fastq"
  Description: "run-folder-archive-fastq"
  Priority: 9
  Data-Types:
    - fastq
- ID: "run-folder-archive-raw"
  Description: "run-folder-archive-raw"
  Priority: 4
- ID: "run-folder-archive-sav"
  Description: "run-folder-archive-sav"
  Priority: 9
  Data-Types:
    - sav
- ID: "run-folder-archive-thumbnails"
  Description: "run-folder-archive-thumbnails"
  Priority: 1
  Data-Types:
    - 
 ~~~~

## Example 3: Archiving large scale scans of museum exhibits


### Requirements

### Our implementation:
0