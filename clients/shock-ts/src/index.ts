export { ShockClient } from "./client.js";
export { ShockError, ShockNetworkError, ShockLockedError } from "./errors.js";
export { smartUpload, chunkedUpload, CHUNK_SIZE } from "./upload.js";
export { isValidNodeId } from "./types.js";

export type {
  ShockResponse,
  ShockPagedResponse,
  PaginatedResult,
  ShockNode,
  ShockFile,
  LockInfo,
  IdxInfo,
  PartsList,
  PartsFile,
  Linkage,
  NodeLocation,
  AclPublic,
  DisplayAcl,
  AclType,
  AnonymousPermissions,
  ShockServerInfo,
  PreAuthResponse,
  ShockClientOptions,
  NodeListQuery,
  UploadOptions,
  DownloadOptions,
  UploadProgress,
  PollOptions,
  ChunkedUploadOptions,
  ChunkedUploadResult,
  SmartUploadOptions,
  LockerState,
  LockedNodes,
  LockedFiles,
  LockedIndexes,
  LocationInfo,
  LocationNodeList,
} from "./types.js";
