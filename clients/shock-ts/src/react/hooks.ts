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
