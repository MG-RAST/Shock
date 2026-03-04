import {
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryResult,
  type UseMutationResult,
} from "@tanstack/react-query";
import { useShockClient } from "./provider.js";
import type {
  DisplayAcl,
  AclType,
  LocationInfo,
  LocationNodeList,
  LockedFiles,
  LockedIndexes,
  LockedNodes,
  LockerState,
  NodeListQuery,
  PaginatedResult,
  ShockNode,
  ShockServerInfo,
  UploadOptions,
} from "../types.js";

// ─── Query Key Factory ───────────────────────────────────────────

export const shockKeys = {
  all: ["shock"] as const,
  serverInfo: () => ["shock", "server-info"] as const,
  nodes: (query?: NodeListQuery) => ["shock", "nodes", query] as const,
  node: (id: string | undefined) => ["shock", "node", id] as const,
  acl: (id: string | undefined) => ["shock", "node", id, "acl"] as const,
  locker: () => ["shock", "locker"] as const,
  lockedNodes: () => ["shock", "locked", "nodes"] as const,
  lockedFiles: () => ["shock", "locked", "files"] as const,
  lockedIndexes: () => ["shock", "locked", "indexes"] as const,
  locationInfo: (locId: string) => ["shock", "location", locId, "info"] as const,
  locationMissing: (locId: string) => ["shock", "location", locId, "missing"] as const,
  locationPresent: (locId: string) => ["shock", "location", locId, "present"] as const,
  traceSummary: () => ["shock", "trace", "summary"] as const,
  traceEvents: () => ["shock", "trace", "events"] as const,
};

// ─── Queries ─────────────────────────────────────────────────────

export function useServerInfo(): UseQueryResult<ShockServerInfo> {
  const client = useShockClient();
  return useQuery({
    queryKey: shockKeys.serverInfo(),
    queryFn: () => client.getServerInfo(),
    staleTime: 5 * 60_000,
  });
}

export function useNode<TAttr = unknown>(
  id: string | undefined
): UseQueryResult<ShockNode<TAttr>> {
  const client = useShockClient();
  return useQuery({
    queryKey: shockKeys.node(id),
    queryFn: () => client.getNode<TAttr>(id!),
    enabled: Boolean(id),
    staleTime: 30_000,
  });
}

export function useNodes<TAttr = unknown>(
  query?: NodeListQuery
): UseQueryResult<PaginatedResult<ShockNode<TAttr>>> {
  const client = useShockClient();
  return useQuery({
    queryKey: shockKeys.nodes(query),
    queryFn: () => client.listNodes<TAttr>(query),
    staleTime: 15_000,
  });
}

export function useNodeAcl(
  id: string | undefined
): UseQueryResult<DisplayAcl> {
  const client = useShockClient();
  return useQuery({
    queryKey: shockKeys.acl(id),
    queryFn: () => client.getAcl(id!),
    enabled: Boolean(id),
  });
}

// ─── Mutations ───────────────────────────────────────────────────

export function useDeleteNode(): UseMutationResult<void, Error, string> {
  const client = useShockClient();
  const qc = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => client.deleteNode(id),
    onSuccess: (_data, id) => {
      qc.removeQueries({ queryKey: shockKeys.node(id) });
      qc.invalidateQueries({ queryKey: ["shock", "nodes"] });
    },
  });
}

export function useUpdateAttributes<TAttr = unknown>(
  id: string
): UseMutationResult<ShockNode<TAttr>, Error, UploadOptions> {
  const client = useShockClient();
  const qc = useQueryClient();

  return useMutation({
    mutationFn: (options: UploadOptions) =>
      client.updateNode<TAttr>(id, options),
    onSuccess: (node) => {
      qc.setQueryData(shockKeys.node(id), node);
      qc.invalidateQueries({ queryKey: ["shock", "nodes"] });
    },
  });
}

export function useAddAcl(
  nodeId: string
): UseMutationResult<
  DisplayAcl,
  Error,
  { type: AclType; users: string[] }
> {
  const client = useShockClient();
  const qc = useQueryClient();

  return useMutation({
    mutationFn: ({ type, users }) => client.addAcl(nodeId, type, users),
    onSuccess: (acl) => {
      qc.setQueryData(shockKeys.acl(nodeId), acl);
    },
  });
}

export function useRemoveAcl(
  nodeId: string
): UseMutationResult<
  DisplayAcl,
  Error,
  { type: AclType; users: string[] }
> {
  const client = useShockClient();
  const qc = useQueryClient();

  return useMutation({
    mutationFn: ({ type, users }) => client.removeAcl(nodeId, type, users),
    onSuccess: (acl) => {
      qc.setQueryData(shockKeys.acl(nodeId), acl);
    },
  });
}

// ─── Index Mutations ──────────────────────────────────────────

export function useCreateIndex(
  nodeId: string
): UseMutationResult<void, Error, string> {
  const client = useShockClient();
  const qc = useQueryClient();

  return useMutation({
    mutationFn: (indexType: string) => client.createIndex(nodeId, indexType),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: shockKeys.node(nodeId) });
    },
  });
}

export function useDeleteIndex(
  nodeId: string
): UseMutationResult<void, Error, string> {
  const client = useShockClient();
  const qc = useQueryClient();

  return useMutation({
    mutationFn: (indexType: string) => client.deleteIndex(nodeId, indexType),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: shockKeys.node(nodeId) });
    },
  });
}

// ─── Admin Queries ────────────────────────────────────────────

export function useLocker(): UseQueryResult<LockerState> {
  const client = useShockClient();
  return useQuery({
    queryKey: shockKeys.locker(),
    queryFn: () => client.getLocker(),
    staleTime: 10_000,
  });
}

export function useLockedNodes(): UseQueryResult<LockedNodes> {
  const client = useShockClient();
  return useQuery({
    queryKey: shockKeys.lockedNodes(),
    queryFn: () => client.getLockedNodes(),
    staleTime: 10_000,
  });
}

export function useLockedFiles(): UseQueryResult<LockedFiles> {
  const client = useShockClient();
  return useQuery({
    queryKey: shockKeys.lockedFiles(),
    queryFn: () => client.getLockedFiles(),
    staleTime: 10_000,
  });
}

export function useLockedIndexes(): UseQueryResult<LockedIndexes> {
  const client = useShockClient();
  return useQuery({
    queryKey: shockKeys.lockedIndexes(),
    queryFn: () => client.getLockedIndexes(),
    staleTime: 10_000,
  });
}

export function useLocationInfo(
  locId: string,
  enabled = true
): UseQueryResult<LocationInfo> {
  const client = useShockClient();
  return useQuery({
    queryKey: shockKeys.locationInfo(locId),
    queryFn: () => client.getLocationInfo(locId),
    enabled,
    staleTime: 60_000,
  });
}

export function useLocationMissing(
  locId: string,
  enabled = true
): UseQueryResult<LocationNodeList> {
  const client = useShockClient();
  return useQuery({
    queryKey: shockKeys.locationMissing(locId),
    queryFn: () => client.getLocationMissing(locId),
    enabled,
    staleTime: 30_000,
  });
}

export function useLocationPresent(
  locId: string,
  enabled = true
): UseQueryResult<LocationNodeList> {
  const client = useShockClient();
  return useQuery({
    queryKey: shockKeys.locationPresent(locId),
    queryFn: () => client.getLocationPresent(locId),
    enabled,
    staleTime: 30_000,
  });
}

// ─── Admin Mutations ──────────────────────────────────────────

export function useStartTrace(): UseMutationResult<string, Error, void> {
  const client = useShockClient();
  return useMutation({
    mutationFn: () => client.startTrace(),
  });
}

export function useStopTrace(): UseMutationResult<string, Error, void> {
  const client = useShockClient();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => client.stopTrace(),
    onSuccess: () => {
      // Invalidate summary/events so they reload after a new trace is captured
      qc.invalidateQueries({ queryKey: shockKeys.traceSummary() });
      qc.invalidateQueries({ queryKey: shockKeys.traceEvents() });
    },
  });
}

export function useTraceSummary(
  enabled = true
): UseQueryResult<string> {
  const client = useShockClient();
  return useQuery({
    queryKey: shockKeys.traceSummary(),
    queryFn: () => client.getTraceSummary(),
    enabled,
    staleTime: 60_000,
  });
}

export function useTraceEvents(
  enabled = true
): UseQueryResult<string> {
  const client = useShockClient();
  return useQuery({
    queryKey: shockKeys.traceEvents(),
    queryFn: () => client.getTraceEvents(),
    enabled,
    staleTime: 60_000,
  });
}

export function useDownloadTrace(): UseMutationResult<Blob, Error, void> {
  const client = useShockClient();
  return useMutation({
    mutationFn: () => client.downloadTrace(),
  });
}
