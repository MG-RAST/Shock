import { useState } from "react";
import {
  useStartTrace,
  useStopTrace,
  useTraceSummary,
  useTraceEvents,
  useDownloadTrace,
} from "shock-client/react";
import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Skeleton } from "@/components/ui/skeleton";
import { Play, Square, Download, Loader2 } from "lucide-react";

export function TraceControls() {
  const startTrace = useStartTrace();
  const stopTrace = useStopTrace();
  const downloadTrace = useDownloadTrace();
  const [showAnalysis, setShowAnalysis] = useState(false);

  const { data: summary, isLoading: summaryLoading, error: summaryError } =
    useTraceSummary(showAnalysis);
  const { data: events, isLoading: eventsLoading, error: eventsError } =
    useTraceEvents(showAnalysis);

  const handleDownload = () => {
    downloadTrace.mutate(undefined, {
      onSuccess: (blob) => {
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = `trace-${Date.now()}.log`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
      },
    });
  };

  const handleStop = () => {
    stopTrace.mutate(undefined, {
      onSuccess: () => {
        // Show analysis tabs after stopping
        setShowAnalysis(true);
      },
    });
  };

  return (
    <div className="space-y-4">
      {/* Controls */}
      <Card>
        <CardHeader>
          <CardTitle className="text-sm">Capture</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Capture a runtime execution trace, then download the raw file or
            view a summary. Raw files can be analyzed locally with{" "}
            <code className="rounded bg-muted px-1 text-xs">go tool trace</code>.
          </p>
          <div className="flex flex-wrap gap-2">
            <Button
              onClick={() => startTrace.mutate()}
              disabled={startTrace.isPending}
            >
              <Play className="mr-2 h-4 w-4" />
              {startTrace.isPending ? "Starting..." : "Start Trace"}
            </Button>
            <Button
              variant="outline"
              onClick={handleStop}
              disabled={stopTrace.isPending}
            >
              <Square className="mr-2 h-4 w-4" />
              {stopTrace.isPending ? "Stopping..." : "Stop Trace"}
            </Button>
            <Button
              variant="outline"
              onClick={handleDownload}
              disabled={downloadTrace.isPending}
            >
              {downloadTrace.isPending ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <Download className="mr-2 h-4 w-4" />
              )}
              Download
            </Button>
          </div>
          {startTrace.data && (
            <p className="text-sm text-green-600 dark:text-green-400">{startTrace.data}</p>
          )}
          {stopTrace.data && (
            <p className="text-sm text-green-600 dark:text-green-400">{stopTrace.data}</p>
          )}
          {downloadTrace.error && (
            <p className="text-sm text-destructive">{downloadTrace.error.message}</p>
          )}
          {(startTrace.error || stopTrace.error) && (
            <p className="text-sm text-destructive">
              {(startTrace.error ?? stopTrace.error)?.message}
            </p>
          )}
        </CardContent>
      </Card>

      {/* Analysis */}
      {!showAnalysis ? (
        <Button variant="secondary" onClick={() => setShowAnalysis(true)}>
          Load trace analysis
        </Button>
      ) : (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">Analysis</CardTitle>
          </CardHeader>
          <CardContent>
            <Tabs defaultValue="summary">
              <TabsList>
                <TabsTrigger value="summary">Footprint Summary</TabsTrigger>
                <TabsTrigger value="events">Parsed Events</TabsTrigger>
              </TabsList>

              <TabsContent value="summary">
                {summaryLoading ? (
                  <Skeleton className="h-48" />
                ) : summaryError ? (
                  <p className="py-4 text-sm text-destructive">
                    {summaryError.message}
                  </p>
                ) : summary ? (
                  <pre className="max-h-[500px] overflow-auto rounded-md bg-muted p-4 text-xs font-mono">
                    {summary}
                  </pre>
                ) : (
                  <p className="py-4 text-sm text-muted-foreground">
                    No trace data. Start and stop a trace first.
                  </p>
                )}
              </TabsContent>

              <TabsContent value="events">
                {eventsLoading ? (
                  <Skeleton className="h-48" />
                ) : eventsError ? (
                  <p className="py-4 text-sm text-destructive">
                    {eventsError.message}
                  </p>
                ) : events ? (
                  <pre className="max-h-[500px] overflow-auto rounded-md bg-muted p-4 text-xs font-mono leading-relaxed">
                    {events}
                  </pre>
                ) : (
                  <p className="py-4 text-sm text-muted-foreground">
                    No trace data. Start and stop a trace first.
                  </p>
                )}
              </TabsContent>
            </Tabs>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
