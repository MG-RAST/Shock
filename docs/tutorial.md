# Getting Started with Shock

This tutorial walks you through running Shock locally and performing common operations using `curl`. By the end you will know how to create nodes, upload and download files, set metadata, query by attributes, and manage access control.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)
- `curl` (comes pre-installed on most systems)
- Optional: [`jq`](https://stedolan.github.io/jq/download/) for pretty-printing JSON responses

## 1. Start Shock

From the repository root:

```bash
docker-compose up -d
```

This starts Shock (port 7445) and MongoDB. Verify the server is running:

```bash
curl http://localhost:7445/ | jq .
```

You should see a JSON response with the Shock version and server information.

## 2. Create Your First Node

Create an empty node (metadata only, no file):

```bash
curl -X POST http://localhost:7445/node | jq .
```

The response contains a `data.id` field -- this is your node's UUID. Save it for later:

```bash
NODE_ID=<paste-your-uuid-here>
```

## 3. Upload a File

Upload a file by creating a new node with a file attached:

```bash
curl -X POST -F 'upload=@myfile.txt' http://localhost:7445/node | jq .
```

The response includes the file name, size, and MD5 checksum under `data.file`. Save the node ID:

```bash
NODE_ID=$(curl -s -X POST -F 'upload=@myfile.txt' http://localhost:7445/node | jq -r .data.id)
echo $NODE_ID
```

You can also upload a file to an existing node:

```bash
curl -X PUT -F 'upload=@myfile.txt' http://localhost:7445/node/$NODE_ID | jq .
```

## 4. Download a File

Download the file stored in a node:

```bash
curl -OJ "http://localhost:7445/node/${NODE_ID}?download"
```

The `-OJ` flags tell curl to save the file using the server-provided filename.

To just stream to stdout:

```bash
curl "http://localhost:7445/node/${NODE_ID}?download"
```

## 5. Set Attributes

Attributes are free-form JSON metadata attached to a node. Add them with an inline string:

```bash
curl -X PUT -F 'attributes_str={"project":"my-experiment", "sample_nr": 1, "organism": "E. coli"}' \
  http://localhost:7445/node/$NODE_ID | jq .
```

Or from a JSON file:

```bash
echo '{"project":"my-experiment", "sample_nr": 1}' > attrs.json
curl -X PUT -F 'attributes=@attrs.json' http://localhost:7445/node/$NODE_ID | jq .
```

## 6. Query Nodes

Search for nodes by attribute values:

```bash
# Find all nodes in a project
curl "http://localhost:7445/node?query&project=my-experiment" | jq .

# Limit results
curl "http://localhost:7445/node?query&project=my-experiment&limit=10" | jq .

# Paginate with offset
curl "http://localhost:7445/node?query&project=my-experiment&limit=10&offset=20" | jq .
```

## 7. View a Node

Retrieve the full metadata for a node:

```bash
curl http://localhost:7445/node/$NODE_ID | jq .
```

## 8. Delete a Node

```bash
curl -X DELETE http://localhost:7445/node/$NODE_ID | jq .
```

## 9. Access Control (ACLs)

When authentication is enabled, nodes have access control lists. View a node's ACLs:

```bash
curl http://localhost:7445/node/$NODE_ID/acl/ | jq .
```

Grant read access to a user:

```bash
curl -X PUT "http://localhost:7445/node/$NODE_ID/acl/read?users=username" | jq .
```

Remove read access:

```bash
curl -X DELETE "http://localhost:7445/node/$NODE_ID/acl/read?users=username" | jq .
```

ACL types: `read`, `write`, `delete`, `owner`.

## 10. S3 Storage with MinIO (Optional)

Shock can store files in S3-compatible object storage. For local development and testing, you can use [MinIO](https://min.io/) as an S3 backend.

### Start the MinIO stack

```bash
docker-compose -f docker-compose.minio.yml up -d shock-mongo shock-minio shock-minio-init shock-server
```

This starts:
- **MinIO** -- S3-compatible object store (API on port 9000, console on port 9001)
- **MongoDB** -- metadata storage
- **Shock** -- configured with auto-upload to MinIO

### Upload a file and verify S3 storage

Upload a file:

```bash
NODE_ID=$(curl -s -X POST -F 'upload=@myfile.txt' http://localhost:7445/node | jq -r .data.id)
echo $NODE_ID
```

Check the node's storage locations (the auto-upload worker copies the file to MinIO in the background):

```bash
# Wait a moment for the auto-upload, then check locations
curl http://localhost:7445/node/$NODE_ID | jq .data.locations
```

You should see an entry for the S3 location once the upload completes.

### Browse MinIO

Open the MinIO console at http://localhost:9001 (login: `minioadmin` / `minioadmin`) to see the uploaded files in the `shock-data` bucket.

### Clean up

```bash
docker-compose -f docker-compose.minio.yml down -v
```

## Next Steps

- [Configuration Guide](./configuration.md) -- customize Shock for your environment
- [API Reference](./API/api.md) -- full REST API documentation
- [Caching and Data Migration](./caching_and_data_migration.md) -- set up multi-tier storage
- [Data Types](./Data-Types.md) -- configure node types and priorities
- [Use Cases](./use_cases.md) -- real-world deployment examples
