import { useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { useUpload } from "shock-client/react";
import { UploadZone } from "@/components/upload/UploadZone";
import { UploadProgress } from "@/components/upload/UploadProgress";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { CheckCircle } from "lucide-react";

export function UploadPage() {
  const navigate = useNavigate();
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [attrsText, setAttrsText] = useState("");
  const { upload, progress, isUploading, error, node, reset } = useUpload();

  const handleFileSelect = useCallback((file: File) => {
    setSelectedFile(file);
  }, []);

  const handleUpload = () => {
    if (!selectedFile) return;

    let attributes: Record<string, unknown> | undefined;
    if (attrsText.trim()) {
      try {
        attributes = JSON.parse(attrsText);
      } catch {
        return; // Don't upload if JSON is invalid
      }
    }

    upload({
      file: selectedFile,
      fileName: selectedFile.name,
      attributes,
    });
  };

  if (node) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">Upload</h1>
        <Card>
          <CardContent className="flex flex-col items-center py-12">
            <CheckCircle className="mb-4 h-12 w-12 text-green-500" />
            <p className="text-lg font-medium">Upload complete!</p>
            <p className="mt-1 text-sm text-muted-foreground font-mono">{node.id}</p>
            <div className="mt-4 flex gap-2">
              <Button onClick={() => navigate(`/ui/nodes/${node.id}`)}>
                View Node
              </Button>
              <Button
                variant="outline"
                onClick={() => {
                  reset();
                  setSelectedFile(null);
                  setAttrsText("");
                }}
              >
                Upload Another
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold">Upload</h1>

      <UploadZone onFileSelect={handleFileSelect} disabled={isUploading} />

      {selectedFile && !isUploading && (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">
              {selectedFile.name}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <label className="mb-1 block text-sm font-medium">
                Attributes (optional JSON)
              </label>
              <textarea
                value={attrsText}
                onChange={(e) => setAttrsText(e.target.value)}
                className="h-24 w-full rounded-md border bg-muted/30 p-3 font-mono text-xs focus:outline-none focus:ring-1 focus:ring-ring"
                placeholder='{"type": "metagenome", "project": "..."}'
                spellCheck={false}
              />
            </div>
            <Button onClick={handleUpload}>Upload File</Button>
          </CardContent>
        </Card>
      )}

      {isUploading && progress && selectedFile && (
        <Card>
          <CardContent className="py-6">
            <UploadProgress progress={progress} fileName={selectedFile.name} />
          </CardContent>
        </Card>
      )}

      {error && (
        <p className="text-sm text-destructive">{error.message}</p>
      )}
    </div>
  );
}
