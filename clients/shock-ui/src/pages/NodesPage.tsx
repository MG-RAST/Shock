import { useState, useCallback } from "react";
import { useNodes, useDeleteNode } from "shock-client/react";
import type { NodeListQuery } from "shock-client";
import { NodeTable } from "@/components/nodes/NodeTable";
import { NodeFilters } from "@/components/nodes/NodeFilters";
import { Pagination } from "@/components/nodes/Pagination";
import { Skeleton } from "@/components/ui/skeleton";

export function NodesPage() {
  const [query, setQuery] = useState<NodeListQuery>({
    limit: 25,
    offset: 0,
    direction: "desc",
    order: "created_on",
  });

  const { data, isLoading, error } = useNodes(query);
  const deleteNode = useDeleteNode();

  const handleSearch = useCallback((q: Record<string, string>) => {
    setQuery((prev) => ({ ...prev, offset: 0, query: Object.keys(q).length > 0 ? q : undefined }));
  }, []);

  const handleSort = useCallback((field: string, direction: "asc" | "desc") => {
    setQuery((prev) => ({ ...prev, order: field, direction }));
  }, []);

  const handleClear = useCallback(() => {
    setQuery((prev) => ({ ...prev, offset: 0, query: undefined }));
  }, []);

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold">Nodes</h1>

      <NodeFilters onSearch={handleSearch} onSort={handleSort} onClear={handleClear} />

      {error && (
        <p className="text-sm text-destructive">{error.message}</p>
      )}

      {isLoading ? (
        <div className="space-y-2">
          {[1, 2, 3, 4, 5].map((i) => <Skeleton key={i} className="h-12" />)}
        </div>
      ) : data ? (
        <>
          <NodeTable
            nodes={data.data}
            onDelete={(id) => deleteNode.mutate(id)}
            deleting={deleteNode.isPending ? deleteNode.variables : undefined}
          />
          <Pagination
            offset={data.offset}
            limit={data.limit}
            totalCount={data.totalCount}
            onOffsetChange={(offset) => setQuery((prev) => ({ ...prev, offset }))}
            onLimitChange={(limit) => setQuery((prev) => ({ ...prev, limit }))}
          />
        </>
      ) : null}
    </div>
  );
}
