import { useLockedNodes, useLockedFiles, useLockedIndexes } from "shock-client/react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Link } from "react-router-dom";

export function LockerView() {
  const { data: nodes, isLoading: nodesLoading } = useLockedNodes();
  const { data: files, isLoading: filesLoading } = useLockedFiles();
  const { data: indexes, isLoading: indexesLoading } = useLockedIndexes();

  if (nodesLoading || filesLoading || indexesLoading) {
    return <div className="space-y-4"><Skeleton className="h-32" /><Skeleton className="h-32" /></div>;
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="text-sm flex items-center gap-2">
            Locked Nodes
            <Badge variant="secondary">{nodes?.length ?? 0}</Badge>
          </CardTitle>
        </CardHeader>
        <CardContent>
          {!nodes || nodes.length === 0 ? (
            <p className="text-sm text-muted-foreground">No locked nodes</p>
          ) : (
            <div className="flex flex-wrap gap-2">
              {nodes.map((id) => (
                <Link key={id} to={`/ui/nodes/${id}`}>
                  <Badge variant="outline" className="font-mono text-xs hover:bg-accent">
                    {id.slice(0, 8)}...
                  </Badge>
                </Link>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm flex items-center gap-2">
            Locked Files
            <Badge variant="secondary">{files ? Object.keys(files).length : 0}</Badge>
          </CardTitle>
        </CardHeader>
        <CardContent>
          {!files || Object.keys(files).length === 0 ? (
            <p className="text-sm text-muted-foreground">No locked files</p>
          ) : (
            <div className="space-y-2">
              {Object.entries(files).map(([key, ids]) => (
                <div key={key}>
                  <span className="text-xs font-medium">{key}:</span>
                  <div className="ml-2 flex flex-wrap gap-1 mt-1">
                    {ids.map((id) => (
                      <Badge key={id} variant="outline" className="font-mono text-xs">
                        {id.slice(0, 8)}...
                      </Badge>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm flex items-center gap-2">
            Locked Indexes
            <Badge variant="secondary">{indexes ? Object.keys(indexes).length : 0}</Badge>
          </CardTitle>
        </CardHeader>
        <CardContent>
          {!indexes || Object.keys(indexes).length === 0 ? (
            <p className="text-sm text-muted-foreground">No locked indexes</p>
          ) : (
            <div className="space-y-2">
              {Object.entries(indexes).map(([key, ids]) => (
                <div key={key}>
                  <span className="text-xs font-medium">{key}:</span>
                  <div className="ml-2 flex flex-wrap gap-1 mt-1">
                    {ids.map((id) => (
                      <Badge key={id} variant="outline" className="font-mono text-xs">
                        {id.slice(0, 8)}...
                      </Badge>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
