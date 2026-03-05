import type { IdxInfo } from "shock-client";
import { useCreateIndex, useDeleteIndex } from "shock-client/react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Plus, Trash2, Lock } from "lucide-react";
import { formatDate } from "@/lib/utils";

interface IndexListProps {
  nodeId: string;
  indexes: Record<string, IdxInfo>;
}

export function IndexList({ nodeId, indexes }: IndexListProps) {
  const createIndex = useCreateIndex(nodeId);
  const deleteIndex = useDeleteIndex(nodeId);
  const [newType, setNewType] = useState("");

  const entries = Object.entries(indexes);

  const handleCreate = () => {
    const trimmed = newType.trim();
    if (!trimmed) return;
    createIndex.mutate(trimmed);
    setNewType("");
  };

  return (
    <div className="space-y-4">
      {entries.length === 0 ? (
        <p className="text-sm text-muted-foreground">No indexes</p>
      ) : (
        <div className="overflow-x-auto rounded-lg border">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="px-4 py-2 text-left font-medium">Type</th>
                <th className="px-4 py-2 text-right font-medium">Units</th>
                <th className="px-4 py-2 text-right font-medium">Avg Size</th>
                <th className="px-4 py-2 text-left font-medium">Created</th>
                <th className="px-4 py-2 text-right font-medium">Actions</th>
              </tr>
            </thead>
            <tbody>
              {entries.map(([type, info]) => (
                <tr key={type} className="border-b">
                  <td className="px-4 py-2 flex items-center gap-2">
                    <span className="font-mono text-xs">{type}</span>
                    {info.locked && (
                      <Badge variant="secondary" className="gap-1">
                        <Lock className="h-3 w-3" /> locked
                      </Badge>
                    )}
                  </td>
                  <td className="px-4 py-2 text-right tabular-nums">{info.total_units}</td>
                  <td className="px-4 py-2 text-right tabular-nums">{info.average_unit_size}</td>
                  <td className="px-4 py-2 text-muted-foreground">{formatDate(info.created_on)}</td>
                  <td className="px-4 py-2 text-right">
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => deleteIndex.mutate(type)}
                      disabled={deleteIndex.isPending}
                      title="Delete index"
                    >
                      <Trash2 className="h-3.5 w-3.5 text-destructive" />
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <div className="flex gap-2">
        <Input
          value={newType}
          onChange={(e) => setNewType(e.target.value)}
          placeholder="Index type (e.g., size, line, chunkrecord)"
          className="h-8 text-sm"
          onKeyDown={(e) => e.key === "Enter" && handleCreate()}
        />
        <Button
          variant="outline"
          size="sm"
          onClick={handleCreate}
          disabled={!newType.trim() || createIndex.isPending}
        >
          <Plus className="mr-1 h-3 w-3" />
          Create
        </Button>
      </div>
    </div>
  );
}
