// ─── Response Envelopes ────────────────────────────────────────────

/** Standard Shock API response envelope (most endpoints). */
export interface ShockResponse<T> {
  status: number;
  data: T | null;
  error: string[] | null;
}

/** Paginated Shock API response envelope (GET /node list). */
export interface ShockPagedResponse<T> extends ShockResponse<T> {
  limit: number;
  offset: number;
  total_count: number;
}

/** Unwrapped paginated result returned by client methods. */
export interface PaginatedResult<T> {
  data: T[];
  limit: number;
  offset: number;
  totalCount: number;
}

// ─── Node ──────────────────────────────────────────────────────────

/**
 * A Shock node. The `TAttr` generic controls the type of `attributes`.
 * Note: ACLs are NEVER included in node JSON (`json:"-"` in Go).
 * Use the `/node/{id}/acl/` endpoint instead.
 */
export interface ShockNode<TAttr = unknown> {
  id: string;
  version: string;
  file: ShockFile;
  attributes: TAttr | null;
  indexes: Record<string, IdxInfo>;
  version_parts: Record<string, string> | null;
  tags: string[] | null;
  linkage: Linkage[] | null;
  priority: number;
  created_on: string;
  last_modified: string;
  expiration: string;
  type: string;
  parts: PartsList | null;
  locations: NodeLocation[] | null;
  restore: boolean;
}

/** File metadata attached to a node. */
export interface ShockFile {
  name: string;
  size: number;
  checksum: Record<string, string>;
  format: string;
  virtual: boolean;
  virtual_parts: string[] | null;
  created_on: string;
  locked: LockInfo | null;
}

/** Lock state for a file being assembled or indexed. */
export interface LockInfo {
  created_on: string;
  error: string;
}

/** Index metadata on a node. */
export interface IdxInfo {
  total_units: number;
  average_unit_size: number;
  created_on: string;
  locked: LockInfo | null;
}

/** A multi-part upload parts list. */
export interface PartsList {
  count: number;
  length: number;
  varlen: boolean;
  parts: PartsFile[] | null;
  compression: string;
}

/** Individual part file record within a PartsList. */
export interface PartsFile {
  name: string;
  size: number;
  checksum: Record<string, string>;
}

/** A linkage between nodes. */
export interface Linkage {
  relation: string;
  ids: string[];
  operation: string;
}

/** Storage location reference on a node. */
export interface NodeLocation {
  id: string;
  stored?: boolean;
  requestedDate?: string;
}

// ─── ACL ───────────────────────────────────────────────────────────

/** Public access flags on an ACL. */
export interface AclPublic {
  read: boolean;
  write: boolean;
  delete: boolean;
}

/** Display ACL returned by `/node/{id}/acl/`. */
export interface DisplayAcl {
  owner: string;
  read: string[];
  write: string[];
  delete: string[];
  public: AclPublic;
}

/** ACL permission types. */
export type AclType = "read" | "write" | "delete" | "owner" | "public_read" | "public_write" | "public_delete";

// ─── Server Info ───────────────────────────────────────────────────

/** Anonymous permission flags from root endpoint. */
export interface AnonymousPermissions {
  read: boolean;
  write: boolean;
  delete: boolean;
}

/** Response from `GET /` (bare struct, NOT envelope-wrapped). */
export interface ShockServerInfo {
  attribute_indexes: string[];
  contact: string;
  id: string;
  auth: string[];
  anonymous_permissions: AnonymousPermissions;
  resources: string[];
  server_time: string;
  type: string;
  url: string;
  uptime: string;
  version: string;
}

// ─── PreAuth ───────────────────────────────────────────────────────

/** Response from `GET /node/{id}?download_url`. */
export interface PreAuthResponse {
  url: string;
  validtill: string;
  format: string;
  filename: string;
  files: number;
  size: number;
}

// ─── Client Options ────────────────────────────────────────────────

export interface ShockClientOptions {
  /** Base URL of the Shock server (e.g. "https://shock.example.com"). */
  url: string;
  /** Static auth token. */
  token?: string;
  /** Dynamic token getter (takes priority over static `token`). */
  getToken?: () => string | undefined;
}

// ─── Query & Upload Options ────────────────────────────────────────

/** Query parameters for listing nodes. */
export interface NodeListQuery {
  limit?: number;
  offset?: number;
  order?: string;
  direction?: "asc" | "desc";
  /** Arbitrary query record fields (e.g. `{ "attributes.type": "metagenome" }`). */
  query?: Record<string, string>;
  /** When true, query against the full node document instead of attributes only. */
  querynode?: boolean;
}

/** Options for creating or updating a node. */
export interface UploadOptions {
  file?: Blob;
  fileName?: string;
  attributes?: Record<string, unknown>;
  expiration?: string;
  parts?: number;
  compression?: "gzip" | "bzip2";
}

/** Options for downloading a file. */
export interface DownloadOptions {
  seek?: number;
  length?: number;
}

/** Upload progress information. */
export interface UploadProgress {
  loaded: number;
  total: number;
  percent: number;
}

/** Options for polling until a node is ready. */
export interface PollOptions {
  intervalMs?: number;
  maxAttempts?: number;
}

/** Options for chunked upload. */
export interface ChunkedUploadOptions {
  file: Blob;
  fileName?: string;
  attributes?: Record<string, unknown>;
  chunkSize?: number;
  onProgress?: (progress: UploadProgress) => void;
  resumeNodeId?: string;
  startPart?: number;
}

/** Result of a chunked upload. */
export interface ChunkedUploadResult<TAttr = unknown> {
  node: ShockNode<TAttr>;
  partsUploaded: number;
}

/** Options for smart upload. */
export interface SmartUploadOptions {
  fileName?: string;
  attributes?: Record<string, unknown>;
  chunkSize?: number;
  onProgress?: (progress: UploadProgress) => void;
}

// ─── Helpers ───────────────────────────────────────────────────────

const NODE_ID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;

/** Validate that a string looks like a Shock node UUID. */
export function isValidNodeId(id: string): boolean {
  return NODE_ID_RE.test(id);
}
