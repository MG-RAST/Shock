import type { ShockClient } from "./client.js";
import type {
  ChunkedUploadOptions,
  ChunkedUploadResult,
  ShockNode,
  SmartUploadOptions,
} from "./types.js";

/** Default chunk size: 2 MiB. */
export const CHUNK_SIZE = 2 * 1024 * 1024;

/**
 * Upload a large file using the Shock multi-part upload protocol.
 *
 * 1. Create a node with `parts=N`
 * 2. Upload each part (1-indexed)
 * 3. Set attributes (also triggers finalization)
 * 4. Poll until assembly is complete
 *
 * Supports resuming: pass `resumeNodeId` and `startPart` to continue
 * from a partial upload. Inspect `node.parts.length` to determine
 * where to resume.
 */
export async function chunkedUpload<TAttr = unknown>(
  client: ShockClient,
  options: ChunkedUploadOptions
): Promise<ChunkedUploadResult<TAttr>> {
  const {
    file,
    fileName,
    attributes,
    chunkSize = CHUNK_SIZE,
    onProgress,
    resumeNodeId,
    startPart,
  } = options;

  const numParts = Math.ceil(file.size / chunkSize);

  // Step 1: Initialize parts node (or use existing one for resume)
  let nodeId: string;
  if (resumeNodeId) {
    nodeId = resumeNodeId;
  } else {
    const node = await client.createNode({ parts: numParts, fileName });
    nodeId = node.id;
  }

  // Step 2: Upload each part (1-indexed!)
  const firstPart = startPart ?? 1;
  let bytesCompleted = (firstPart - 1) * chunkSize;

  for (let i = firstPart; i <= numParts; i++) {
    const start = (i - 1) * chunkSize;
    const end = Math.min(start + chunkSize, file.size);
    const chunk = file.slice(start, end);

    await client.uploadPart(nodeId, i, chunk, (partProgress) => {
      onProgress?.({
        loaded: bytesCompleted + partProgress.loaded,
        total: file.size,
        percent: Math.round(
          ((bytesCompleted + partProgress.loaded) / file.size) * 100
        ),
      });
    });
    bytesCompleted = end;
  }

  // Step 3: Set attributes (also serves as finalization signal)
  await client.updateNode(nodeId, { attributes: attributes ?? {} });

  // Step 4: Poll until assembly complete
  // The server auto-assembles when all parts are received.
  // file.locked will be non-null during assembly.
  const finalNode = await client.pollUntilReady<TAttr>(nodeId);

  return { node: finalNode, partsUploaded: numParts };
}

/**
 * Automatically choose between single-request upload and chunked upload
 * based on file size.
 */
export async function smartUpload<TAttr = unknown>(
  client: ShockClient,
  file: Blob,
  options?: SmartUploadOptions
): Promise<ShockNode<TAttr>> {
  const chunkSize = options?.chunkSize ?? CHUNK_SIZE;

  if (file.size <= chunkSize) {
    // Simple single-request upload
    return client.createNode<TAttr>({
      file,
      fileName: options?.fileName,
      attributes: options?.attributes,
    });
  }

  const result = await chunkedUpload<TAttr>(client, {
    file,
    fileName: options?.fileName,
    attributes: options?.attributes,
    chunkSize,
    onProgress: options?.onProgress,
  });
  return result.node;
}
