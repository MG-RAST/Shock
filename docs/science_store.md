# Building a data store for scientific data


## Example 1: MG-RAST
### requirements:
- ~ 1PB of files
- indexing
- automated deletion of intermediary results (expiry) after set intervals

### Our implementation:

- set up a dedicated Server with initially 1PB of local storage on commodity hardware
- added an S3 backend
- added a TSM backend store
- set up node removal from the main store IF a node is stored in two locations (S3 and TSM)


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
