# shock-client

TypeScript client for [Shock](https://github.com/MG-RAST/Shock) object storage. Two layers:

1. **Core** (`shock-client`) — framework-agnostic `ShockClient` class, works in Node.js and browsers
2. **React** (`shock-client/react`) — hooks built on [TanStack Query](https://tanstack.com/query) v5

## Installation

```bash
npm install shock-client
```

For React hooks, also install peer dependencies:

```bash
npm install react @tanstack/react-query
```

## Quick Start

### Core Client

```typescript
import { ShockClient } from "shock-client";

const client = new ShockClient({
  url: "https://shock.example.com",
  token: "your-oauth-token",
});

// Create a node with attributes
const node = await client.createNode({
  attributes: { type: "metagenome", project: "demo" },
});

// Upload a file
const fileNode = await client.createNode({
  file: someBlob,
  fileName: "sample.fastq",
  attributes: { format: "fastq" },
});

// Get a node
const fetched = await client.getNode(node.id);

// List nodes with query
const results = await client.listNodes({
  limit: 25,
  query: { "attributes.type": "metagenome" },
});
console.log(results.data, results.totalCount);

// Download
const blob = await client.downloadFile(node.id);
```

### Dynamic Auth Token

Pass a `getToken` function for integration with token refresh flows:

```typescript
const client = new ShockClient({
  url: "https://shock.example.com",
  getToken: () => authStore.getAccessToken(),
});
```

### React Hooks

```tsx
import { ShockProvider } from "shock-client/react";

function App() {
  return (
    <ShockProvider url="https://shock.example.com" token={authToken}>
      <NodeList />
    </ShockProvider>
  );
}
```

```tsx
import { useNodes, useNode, useDeleteNode } from "shock-client/react";

function NodeList() {
  const { data, isLoading } = useNodes({ limit: 20 });
  const deleteNode = useDeleteNode();

  if (isLoading) return <p>Loading...</p>;

  return (
    <ul>
      {data?.data.map((node) => (
        <li key={node.id}>
          {node.id} — {node.file.name}
          <button onClick={() => deleteNode.mutate(node.id)}>Delete</button>
        </li>
      ))}
    </ul>
  );
}
```

## Core API Reference

### `ShockClient`

#### Constructor

```typescript
new ShockClient(options: {
  url: string;                          // Shock server base URL
  token?: string;                       // Static OAuth token
  getToken?: () => string | undefined;  // Dynamic token getter (takes priority)
})
```

#### Node Methods

| Method | Description |
|--------|-------------|
| `getNode<TAttr>(id)` | Fetch a single node |
| `listNodes<TAttr>(query?)` | List nodes with pagination and filtering |
| `createNode<TAttr>(options?)` | Create a node (with optional file/attributes) |
| `updateNode<TAttr>(id, options)` | Update node attributes or upload a file |
| `deleteNode(id)` | Delete a node |

#### File Methods

| Method | Description |
|--------|-------------|
| `uploadPart<TAttr>(nodeId, partNumber, data, onProgress?)` | Upload a single part (1-indexed) |
| `downloadFile(id, options?)` | Download file as Blob, with optional `seek`/`length` |
| `getDownloadUrl(id)` | Get a pre-authenticated download URL |
| `pollUntilReady<TAttr>(nodeId, options?)` | Poll until file assembly/indexing completes |

#### ACL Methods

| Method | Description |
|--------|-------------|
| `getAcl(nodeId)` | Get node access control list |
| `addAcl(nodeId, type, users)` | Grant access to users |
| `removeAcl(nodeId, type, users)` | Revoke access from users |

ACL types: `"read"`, `"write"`, `"delete"`, `"owner"`, `"public_read"`, `"public_write"`, `"public_delete"`

#### Other Methods

| Method | Description |
|--------|-------------|
| `getServerInfo()` | Get server version and configuration |
| `createIndex(nodeId, indexType)` | Create or rebuild an index |
| `getLocations(nodeId)` | List storage locations for a node |
| `setToken(token?)` | Update the static auth token |

### Typed Attributes

All node methods accept a `TAttr` generic for typed attributes:

```typescript
interface MyAttributes {
  project: string;
  sample_id: number;
}

const node = await client.getNode<MyAttributes>("abc-123-...");
node.attributes?.project; // string
```

### Uploads

#### Simple Upload

Files under the chunk size threshold (default 2 MiB) upload in a single request:

```typescript
import { smartUpload } from "shock-client";

const node = await smartUpload(client, file, {
  fileName: "data.fastq.gz",
  attributes: { format: "fastq" },
  onProgress: ({ percent }) => console.log(`${percent}%`),
});
```

#### Chunked Upload

Large files are automatically split into parts:

```typescript
import { chunkedUpload, CHUNK_SIZE } from "shock-client";

const result = await chunkedUpload(client, {
  file: largeFile,
  fileName: "big-assembly.fasta",
  chunkSize: 10 * 1024 * 1024, // 10 MiB parts
  onProgress: ({ loaded, total, percent }) => {
    console.log(`${loaded}/${total} bytes (${percent}%)`);
  },
});

console.log(result.node.id, result.partsUploaded);
```

#### Resuming Uploads

Pass `resumeNodeId` and `startPart` to continue a failed upload:

```typescript
const result = await chunkedUpload(client, {
  file: largeFile,
  resumeNodeId: "existing-node-uuid",
  startPart: 5, // resume from part 5
});
```

Upload progress callbacks use `XMLHttpRequest` in browser environments. Node.js environments get no per-part progress (only per-part completion).

## React Hooks Reference

### `<ShockProvider>`

Wraps your app with a `ShockClient` and `QueryClientProvider`:

```tsx
<ShockProvider
  url="https://shock.example.com"
  token={token}                    // or getToken={() => ...}
  queryClient={existingQC}         // optional, creates one internally otherwise
>
  {children}
</ShockProvider>
```

Token changes call `client.setToken()` instead of re-creating the client, preserving the query cache.

### Query Hooks

| Hook | Returns | Notes |
|------|---------|-------|
| `useServerInfo()` | `UseQueryResult<ShockServerInfo>` | Stale after 5 min |
| `useNode(id)` | `UseQueryResult<ShockNode<TAttr>>` | Disabled when `id` is falsy; stale after 30s |
| `useNodes(query?)` | `UseQueryResult<PaginatedResult<ShockNode>>` | Stale after 15s |
| `useNodeAcl(id)` | `UseQueryResult<DisplayAcl>` | Disabled when `id` is falsy |

### Mutation Hooks

| Hook | Input | Notes |
|------|-------|-------|
| `useDeleteNode()` | `mutate(id)` | Removes node from cache, invalidates lists |
| `useUpdateAttributes(id)` | `mutate(options)` | Optimistically updates node cache |
| `useAddAcl(nodeId)` | `mutate({ type, users })` | Updates ACL cache |
| `useRemoveAcl(nodeId)` | `mutate({ type, users })` | Updates ACL cache |

### `useUpload`

Upload hook with progress tracking:

```tsx
import { useUpload } from "shock-client/react";

function Uploader() {
  const { upload, progress, isUploading, error, node, reset } = useUpload();

  return (
    <div>
      <input
        type="file"
        onChange={(e) => {
          const file = e.target.files?.[0];
          if (file) {
            upload({
              file,
              fileName: file.name,
              attributes: { uploaded_by: "web" },
            });
          }
        }}
      />
      {isUploading && progress && <p>{progress.percent}% uploaded</p>}
      {error && <p>Error: {error.message}</p>}
      {node && <p>Uploaded: {node.id}</p>}
    </div>
  );
}
```

### Query Key Factory

Use `shockKeys` for manual cache operations:

```typescript
import { shockKeys } from "shock-client/react";

// Invalidate all Shock queries
queryClient.invalidateQueries({ queryKey: shockKeys.all });

// Invalidate a specific node
queryClient.invalidateQueries({ queryKey: shockKeys.node(nodeId) });
```

## Error Handling

Three error classes, all exported from `shock-client`:

| Class | When | Properties |
|-------|------|------------|
| `ShockError` | Server returned a non-2xx response | `status`, `messages` |
| `ShockNetworkError` | Fetch/XHR failed (DNS, timeout, etc.) | `networkCause` |
| `ShockLockedError` | HTTP 423 — file locked during assembly | `status`, `messages` |

```typescript
import { ShockError, ShockLockedError } from "shock-client";

try {
  await client.getNode(id);
} catch (err) {
  if (err instanceof ShockLockedError) {
    // Retry later — file is being assembled
  } else if (err instanceof ShockError) {
    console.error(`HTTP ${err.status}:`, err.messages);
  }
}
```

## Important Notes

- **CORS**: All mutations use `FormData` (never JSON bodies) because the Shock server only allows `Authorization` in CORS headers. Sending `Content-Type: application/json` triggers a preflight that the server rejects.
- **ACLs are separate**: Node JSON never includes ACLs. Always use `getAcl()` / `useNodeAcl()`.
- **File immutability**: Once a node has a file, it cannot be replaced.
- **Part numbering**: Upload parts are 1-indexed. The client validates this and throws if you pass 0 or negative numbers.
- **Node ID validation**: All methods validate node IDs as UUIDs to prevent path traversal.

## Requirements

- Node.js >= 18 (uses native `fetch`)
- React >= 18 and `@tanstack/react-query` >= 5 (for hooks only, optional peer deps)

## Building from Source

```bash
cd clients/shock-ts
npm install
npm run build      # ESM + CJS + .d.ts via tsup
npm run typecheck   # Type-check without emitting
```

## License

BSD-3-Clause
