import { useState } from "react";
import { useShockClient } from "shock-client/react";
import { Button } from "@/components/ui/button";
import { Link2, Check, Loader2 } from "lucide-react";

interface PreAuthButtonProps {
  nodeId: string;
}

export function PreAuthButton({ nodeId }: PreAuthButtonProps) {
  const client = useShockClient();
  const [url, setUrl] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [copied, setCopied] = useState(false);

  const handleGetUrl = async () => {
    setLoading(true);
    try {
      const result = await client.getDownloadUrl(nodeId);
      setUrl(result.url);
    } catch (err) {
      console.error("Failed to get download URL:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleCopy = async () => {
    if (!url) return;
    await navigator.clipboard.writeText(url);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  if (url) {
    return (
      <div className="space-y-2">
        <div className="flex gap-2">
          <input
            readOnly
            value={url}
            className="flex-1 rounded-md border bg-muted px-3 py-2 text-xs font-mono"
          />
          <Button variant="outline" size="sm" onClick={handleCopy}>
            {copied ? <Check className="h-4 w-4" /> : "Copy"}
          </Button>
        </div>
      </div>
    );
  }

  return (
    <Button variant="outline" onClick={handleGetUrl} disabled={loading}>
      {loading ? (
        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
      ) : (
        <Link2 className="mr-2 h-4 w-4" />
      )}
      Pre-Auth URL
    </Button>
  );
}
