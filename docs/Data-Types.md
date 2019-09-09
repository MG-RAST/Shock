# NEEDS REWRITING FOR NEW TYPES INFRASTRUCTURE



These data types are only recommendations. Shock does no validation.

## Type "data-library"
Example:
```bash
"attributes": {
      "type": "data-library",
      "name": "Solr M5NR",
      "version": "1",
      "description": "Solr M5NR v1 with Solr v4.10.3",
      "member": "1/1",
      "project": "production",
      "provenance": {
        "creation_type": "manual",
        "note": "tar -zcvf solr-m5nr_v1_solr_v4.10.3.tgz -C /mnt/m5nr_1/data/index/ ."
      }   
}
```


Required fields:<br>
**type=data-library** Application Scope/Name<br>
**name=\<string>** (e.g. “M5NR”, or “Bowtie index of human genome”)<br>
**version=\<string>** (version number, date or similar of the data-library-name)<br>

Optional fields:<br>
**member=\<string>** (a name for the data library member, could be the same as filename, or chunk number, e.g. “m5nr.1”)<br>

Filename is stored under Shock metadata **file->name** and is not part of this specification.<br>


**description=\<string>** (longer description)<br>
**file_format=\<string>** (fasta, bt2 ... etc., CV would be nice long term)<br>
**created=\<date>** (creation date of the file/member, not upload date; is this provenance?)<br>

**attributes->provenance->creation_type = clone | workflow | manual**<br>

1) clone<br>
simple case: data just has been copied from another server (e.g. BLAST NR)<br>
**attributes->provenance->url=<url>** URI for the original file , download/copy location<br>

2) AWE workflow<br>
**attributes->provenance->workflow=\<url/string>** Reference to a workflow document if available, if this a computed product and not copied (workflow document with input)<br>

3) manual<br>
**attributes->provenance->note=\<string>** Description how the file has been created if not copied/downloaded<br>

### Comment
Every “data-library” consists of a finite number of Shock nodes. A library name AND version number uniquely identifies a specific library. (Unsolved: how do you prevent people uploading stuff with the same name? protected namespaces?)

## Type "dockerimage"
TODO

## Type \<other>
TODO