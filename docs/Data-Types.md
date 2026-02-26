# Shock Data Types

Shock uses a YAML-based type system to classify nodes. Types are defined in a `Types.yaml` configuration file that is loaded at server startup.

## Types.yaml Format

```yaml
Types:
  - ID: "type_id"
    Description: "Human-readable description"
    Priority: 0
    Data-Types:
      - extension1
      - extension2
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `ID` | string | Unique identifier for the type (e.g. `"metagenome"`, `"temp"`, `"default"`) |
| `Description` | string | Human-readable description of the type |
| `Priority` | int | Priority value (0 = lowest, 9+ = highest). Used by the migration system to determine which nodes are eligible for remote storage. Locations with a `MinPriority` setting will only accept nodes at or above that priority. |
| `Data-Types` | list | Optional list of file extensions associated with this type (e.g. `fastq`, `fasta`, `bam`) |

## How Types Work

- Each node in Shock has a `type` field that references a type ID from `Types.yaml`.
- The default type is `"basic"` if no type is specified at node creation.
- The `Priority` field is central to the data migration system: it determines which nodes get migrated to remote locations. For example, a location with `MinPriority: 7` will only accept nodes whose type has `Priority >= 7`, preventing temporary or low-value files from being stored in expensive remote storage.
- The `Data-Types` list is informational and describes what file formats are expected for nodes of this type.

## Example Types.yaml

This example is from `test/config.d-minio/Types.yaml`:

```yaml
Types:
  - ID: "default"
    Description: "default"
    Priority: 0
  - ID: "temp"
    Description: "temporary file"
    Priority: 0
  - ID: "VM"
    Description: "Virtual Machine"
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
  - ID: "cv"
    Description: "Controlled Vocabulary"
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
```

## Querying Types via API

The `/types` API endpoint provides information about configured types:

```bash
curl -s http://localhost:7445/types/mixs/info | jq .
```

```json
{
  "status": 200,
  "data": {
    "id": "mixs",
    "description": "GSC MIxS Metadata file XLS format",
    "priority": 9
  },
  "error": null
}
```

See the [Types API documentation](./API/types.md) for all available endpoints.

## Legacy Attribute-Based Types

Older Shock deployments may use attribute-based types where the type is stored in node attributes rather than the dedicated type field. These data types are conventions rather than enforced schemas -- Shock does no validation of attribute content.

### Type "data-library"

```json
{
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
}
```

Required fields:
- **type=data-library** -- Application scope/name
- **name=\<string>** -- e.g. "M5NR" or "Bowtie index of human genome"
- **version=\<string>** -- Version number, date, or similar

Optional fields:
- **member=\<string>** -- Name for the data library member (e.g. chunk number)
- **description=\<string>** -- Longer description
- **file_format=\<string>** -- File format (fasta, bt2, etc.)
- **provenance** -- Object describing how the data was created (`clone`, `workflow`, or `manual`)
