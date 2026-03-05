import { useServerInfo, useLockedNodes, useLockedFiles, useLockedIndexes } from "shock-client/react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Server, Lock, Clock } from "lucide-react";

export function Dashboard() {
  const { data: info, isLoading: infoLoading } = useServerInfo();
  const { data: lockedNodes } = useLockedNodes();
  const { data: lockedFiles } = useLockedFiles();
  const { data: lockedIndexes } = useLockedIndexes();

  if (infoLoading) {
    return (
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {[1, 2, 3].map((i) => <Skeleton key={i} className="h-32" />)}
      </div>
    );
  }

  if (!info) return null;

  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center gap-3 pb-2">
            <Server className="h-5 w-5 text-muted-foreground" />
            <CardTitle className="text-sm font-medium">Server</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-1 text-sm">
              <p>Version: <span className="font-mono">{info.version}</span></p>
              <p className="flex items-center gap-1">
                <Clock className="h-3 w-3" /> Uptime: {info.uptime}
              </p>
              <p>Contact: {info.contact || "—"}</p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center gap-3 pb-2">
            <Lock className="h-5 w-5 text-muted-foreground" />
            <CardTitle className="text-sm font-medium">Locked Resources</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-1 text-sm">
              <p>Nodes: <Badge variant="secondary">{lockedNodes?.length ?? "—"}</Badge></p>
              <p>Files: <Badge variant="secondary">{lockedFiles ? Object.keys(lockedFiles).length : "—"}</Badge></p>
              <p>Indexes: <Badge variant="secondary">{lockedIndexes ? Object.keys(lockedIndexes).length : "—"}</Badge></p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Configuration</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-1 text-sm">
              <p>Auth: {info.auth && info.auth.length > 0 ? info.auth.join(", ") : "basic"}</p>
              <div className="flex gap-1">
                <span>Anon:</span>
                {info.anonymous_permissions.read && <Badge variant="outline">read</Badge>}
                {info.anonymous_permissions.write && <Badge variant="outline">write</Badge>}
                {info.anonymous_permissions.delete && <Badge variant="outline">delete</Badge>}
                {!info.anonymous_permissions.read && !info.anonymous_permissions.write && !info.anonymous_permissions.delete && (
                  <span className="text-muted-foreground">none</span>
                )}
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
