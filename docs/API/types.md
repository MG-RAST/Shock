

### API Routes for /types (default port 7445):

(TODO: move this into openapi) 


##### OPTIONS

- Permitted by everyone:
  - all options requests respond with CORS headers and 200 OK

##### POST
- Permitted by everyone:
- N/A

- Permitted by admis 
  - `/types/<type_id>/info`  provide info on a data type
  
##### GET

- Permitted by everyone:
 - N/A

- Permitted by admis 
   


##### PUT

- N/A

Note: Configurations is via the Types.yaml file at server start time.


##### DELETE

- N/A

<br>


## An overview of /types

~~~~
>curl -s -X GET "localhost:7445/types/mixs/info" | jq .
{
  "status": 200,
  "data": {
    "id": "mixs",
    "description": "GSC MIxS Metadata file XLS format",
    "priority": 9
  },
  "error": null
}
~~~~



### Type.yaml from the configuration

~~~~
Types:
- ID: "default"
  Description: "default"
  Priority: 0
- ID: "temp"
  Description: "temporary file"
  Priority: 0
- ID: "EBI Submission Receipt"
  Description: "EBI Submission Receipt"
  Priority: 7
- ID: "VM"
  Description: "Virtual Machine"
  Priority: 1
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
- ID: "image"
  Description: "image file"
  Priority: 1
  Data-Types:
    - jpeg
    - jpg
    - gif
    - tif
    - png
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


