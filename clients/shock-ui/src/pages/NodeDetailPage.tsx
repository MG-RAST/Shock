import { useParams, Link } from "react-router-dom";
import { useNode } from "shock-client/react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { DownloadButton } from "@/components/download/DownloadButton";
import { PreAuthButton } from "@/components/download/PreAuthButton";
import { AclPanel } from "@/components/acl/AclPanel";
import { IndexList } from "@/components/indexes/IndexList";
import { AttributeEditor } from "@/components/nodes/AttributeEditor";
import { formatBytes, formatDate } from "@/lib/utils";
import { ArrowLeft } from "lucide-react";

export function NodeDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { data: node, isLoading, error } = useNode(id);

  if (isLoading) {
    return <div className="space-y-4"><Skeleton className="h-8 w-48" /><Skeleton className="h-64" /></div>;
  }

  if (error) {
    return <p className="text-destructive">{error.message}</p>;
  }

  if (!node || !id) return null;

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link to="/ui/nodes" className="text-muted-foreground hover:text-foreground">
          <ArrowLeft className="h-5 w-5" />
        </Link>
        <div>
          <h1 className="text-2xl font-bold">{node.file.name || "(unnamed)"}</h1>
          <p className="font-mono text-xs text-muted-foreground">{node.id}</p>
        </div>
        {node.file.locked && <Badge variant="secondary">locked</Badge>}
      </div>

      {/* File info card */}
      <Card>
        <CardHeader>
          <CardTitle className="text-sm">File Info</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <div>
              <p className="text-xs text-muted-foreground">Size</p>
              <p className="font-mono">{node.file.size > 0 ? formatBytes(node.file.size) : "—"}</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Format</p>
              <p>{node.file.format || "—"}</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Created</p>
              <p>{formatDate(node.created_on)}</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Modified</p>
              <p>{formatDate(node.last_modified)}</p>
            </div>
          </div>

          {node.file.checksum && Object.keys(node.file.checksum).length > 0 && (
            <div className="mt-4">
              <p className="text-xs text-muted-foreground mb-1">Checksums</p>
              <div className="space-y-1">
                {Object.entries(node.file.checksum).map(([alg, hash]) => (
                  <div key={alg} className="flex gap-2">
                    <Badge variant="outline" className="text-xs">{alg}</Badge>
                    <span className="font-mono text-xs break-all">{hash}</span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {node.file.size > 0 && (
            <div className="mt-4 flex gap-2">
              <DownloadButton nodeId={id} fileName={node.file.name} />
              <PreAuthButton nodeId={id} />
            </div>
          )}
        </CardContent>
      </Card>

      {/* Tabs for attributes, ACL, indexes, locations */}
      <Tabs defaultValue="attributes">
        <TabsList>
          <TabsTrigger value="attributes">Attributes</TabsTrigger>
          <TabsTrigger value="acl">ACL</TabsTrigger>
          <TabsTrigger value="indexes">
            Indexes ({Object.keys(node.indexes).length})
          </TabsTrigger>
          <TabsTrigger value="locations">
            Locations ({node.locations?.length ?? 0})
          </TabsTrigger>
        </TabsList>

        <TabsContent value="attributes">
          <Card>
            <CardContent className="pt-6">
              <AttributeEditor nodeId={id} attributes={node.attributes} />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="acl">
          <Card>
            <CardContent className="pt-6">
              <AclPanel nodeId={id} />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="indexes">
          <Card>
            <CardContent className="pt-6">
              <IndexList nodeId={id} indexes={node.indexes} />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="locations">
          <Card>
            <CardContent className="pt-6">
              {node.locations && node.locations.length > 0 ? (
                <div className="overflow-x-auto rounded-lg border">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b bg-muted/50">
                        <th className="px-4 py-2 text-left font-medium">ID</th>
                        <th className="px-4 py-2 text-left font-medium">Stored</th>
                        <th className="px-4 py-2 text-left font-medium">Requested</th>
                      </tr>
                    </thead>
                    <tbody>
                      {node.locations.map((loc) => (
                        <tr key={loc.id} className="border-b">
                          <td className="px-4 py-2 font-mono text-xs">{loc.id}</td>
                          <td className="px-4 py-2">
                            <Badge variant={loc.stored ? "default" : "secondary"}>
                              {loc.stored ? "yes" : "pending"}
                            </Badge>
                          </td>
                          <td className="px-4 py-2 text-muted-foreground">
                            {loc.requestedDate ?? "—"}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <p className="text-sm text-muted-foreground">No storage locations</p>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
