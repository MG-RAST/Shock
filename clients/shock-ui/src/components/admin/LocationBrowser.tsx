import { useState } from "react";
import { useLocationInfo, useLocationMissing, useLocationPresent } from "shock-client/react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Search } from "lucide-react";

export function LocationBrowser() {
  const [locId, setLocId] = useState("");
  const [activeId, setActiveId] = useState<string | null>(null);

  const { data: info, isLoading: infoLoading, error: infoError } = useLocationInfo(
    activeId ?? "",
    Boolean(activeId)
  );
  const { data: missing } = useLocationMissing(activeId ?? "", Boolean(activeId));
  const { data: present } = useLocationPresent(activeId ?? "", Boolean(activeId));

  const handleSearch = () => {
    if (locId.trim()) setActiveId(locId.trim());
  };

  return (
    <div className="space-y-4">
      <div className="flex gap-2">
        <Input
          value={locId}
          onChange={(e) => setLocId(e.target.value)}
          placeholder="Enter location ID (e.g., s3)"
          onKeyDown={(e) => e.key === "Enter" && handleSearch()}
        />
        <Button onClick={handleSearch} disabled={!locId.trim()}>
          <Search className="mr-2 h-4 w-4" />
          Inspect
        </Button>
      </div>

      {activeId && (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">Location: {activeId}</CardTitle>
          </CardHeader>
          <CardContent>
            {infoLoading ? (
              <Skeleton className="h-20" />
            ) : infoError ? (
              <p className="text-sm text-destructive">
                {infoError instanceof Error ? infoError.message : "Failed to load location info"}
              </p>
            ) : info ? (
              <div className="space-y-4">
                <pre className="rounded-md bg-muted p-3 text-xs font-mono overflow-auto max-h-40">
                  {JSON.stringify(info, null, 2)}
                </pre>

                <Tabs defaultValue="missing">
                  <TabsList>
                    <TabsTrigger value="missing">
                      Missing <Badge variant="secondary" className="ml-1">{missing?.length ?? "—"}</Badge>
                    </TabsTrigger>
                    <TabsTrigger value="present">
                      Present <Badge variant="secondary" className="ml-1">{present?.length ?? "—"}</Badge>
                    </TabsTrigger>
                  </TabsList>
                  <TabsContent value="missing">
                    {!missing || missing.length === 0 ? (
                      <p className="text-sm text-muted-foreground py-2">No missing nodes</p>
                    ) : (
                      <div className="max-h-48 overflow-auto">
                        <div className="flex flex-wrap gap-1 py-2">
                          {missing.map((id) => (
                            <Badge key={id} variant="outline" className="font-mono text-xs">
                              {id.slice(0, 8)}...
                            </Badge>
                          ))}
                        </div>
                      </div>
                    )}
                  </TabsContent>
                  <TabsContent value="present">
                    {!present || present.length === 0 ? (
                      <p className="text-sm text-muted-foreground py-2">No present nodes</p>
                    ) : (
                      <div className="max-h-48 overflow-auto">
                        <div className="flex flex-wrap gap-1 py-2">
                          {present.map((id) => (
                            <Badge key={id} variant="outline" className="font-mono text-xs">
                              {id.slice(0, 8)}...
                            </Badge>
                          ))}
                        </div>
                      </div>
                    )}
                  </TabsContent>
                </Tabs>
              </div>
            ) : null}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
