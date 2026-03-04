import { Link } from "react-router-dom";
import type { ShockNode } from "shock-client";
import { formatBytes, formatDate } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { FileText, Trash2, Download } from "lucide-react";

interface NodeTableProps {
  nodes: ShockNode[];
  onDelete: (id: string) => void;
  deleting?: string;
}

export function NodeTable({ nodes, onDelete, deleting }: NodeTableProps) {
  if (nodes.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
        <FileText className="mb-2 h-8 w-8" />
        <p>No nodes found</p>
      </div>
    );
  }

  return (
    <div className="overflow-x-auto rounded-lg border">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b bg-muted/50">
            <th className="px-4 py-3 text-left font-medium">Name</th>
            <th className="px-4 py-3 text-left font-medium">ID</th>
            <th className="px-4 py-3 text-right font-medium">Size</th>
            <th className="px-4 py-3 text-left font-medium">Format</th>
            <th className="px-4 py-3 text-left font-medium">Created</th>
            <th className="px-4 py-3 text-right font-medium">Actions</th>
          </tr>
        </thead>
        <tbody>
          {nodes.map((node) => (
            <tr key={node.id} className="border-b transition-colors hover:bg-muted/30">
              <td className="px-4 py-3">
                <Link
                  to={`/ui/nodes/${node.id}`}
                  className="font-medium text-primary hover:underline"
                >
                  {node.file.name || "(unnamed)"}
                </Link>
                {node.file.locked && (
                  <Badge variant="secondary" className="ml-2">locked</Badge>
                )}
              </td>
              <td className="px-4 py-3 font-mono text-xs text-muted-foreground">
                {node.id.slice(0, 8)}...
              </td>
              <td className="px-4 py-3 text-right tabular-nums">
                {node.file.size > 0 ? formatBytes(node.file.size) : "-"}
              </td>
              <td className="px-4 py-3">
                {node.file.format ? (
                  <Badge variant="outline">{node.file.format}</Badge>
                ) : "-"}
              </td>
              <td className="px-4 py-3 text-muted-foreground">
                {formatDate(node.created_on)}
              </td>
              <td className="px-4 py-3 text-right">
                <div className="flex items-center justify-end gap-1">
                  {node.file.size > 0 && (
                    <Link to={`/ui/nodes/${node.id}`}>
                      <Button variant="ghost" size="icon" title="Download">
                        <Download className="h-3.5 w-3.5" />
                      </Button>
                    </Link>
                  )}
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => onDelete(node.id)}
                    disabled={deleting === node.id}
                    title="Delete"
                  >
                    <Trash2 className="h-3.5 w-3.5 text-destructive" />
                  </Button>
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
