import { useState, useCallback } from "react";
import { useServerInfo } from "shock-client/react";
import { useAuth } from "@/hooks/use-auth";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Check, Copy, Sun, Moon, Shield, User, Terminal } from "lucide-react";

const PAGE_SIZE_KEY = "shock_page_size";

export function getStoredPageSize(): number {
  const stored = localStorage.getItem(PAGE_SIZE_KEY);
  const parsed = stored ? Number(stored) : 25;
  return Number.isFinite(parsed) ? parsed : 25;
}

export function setStoredPageSize(size: number): void {
  localStorage.setItem(PAGE_SIZE_KEY, String(size));
}

function CopyField({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(async () => {
    await navigator.clipboard.writeText(value);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }, [value]);

  return (
    <div className="space-y-1">
      <label className="text-xs font-medium text-muted-foreground">{label}</label>
      <div className="flex gap-2">
        <Input
          readOnly
          value={value}
          className={`flex-1 ${mono ? "font-mono text-xs" : ""}`}
          onClick={(e) => (e.target as HTMLInputElement).select()}
        />
        <Button variant="outline" size="sm" onClick={handleCopy} className="shrink-0">
          {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
        </Button>
      </div>
    </div>
  );
}

export function SettingsPage() {
  const { token, username, isAdmin } = useAuth();
  const { data: serverInfo } = useServerInfo();
  const [dark, setDark] = useState(() =>
    document.documentElement.classList.contains("dark")
  );
  const [pageSize, setPageSize] = useState(getStoredPageSize);

  const toggleTheme = useCallback(() => {
    setDark((prev) => {
      const next = !prev;
      document.documentElement.classList.toggle("dark", next);
      return next;
    });
  }, []);

  const handlePageSizeChange = useCallback((size: number) => {
    setPageSize(size);
    setStoredPageSize(size);
  }, []);

  const authHeader = token ? `Basic ${token}` : "";
  const serverUrl = serverInfo?.url?.replace(/\/$/, "") || window.location.origin;
  const curlExample = `curl -H "Authorization: ${authHeader}" ${serverUrl}/node?limit=10`;

  return (
    <div className="space-y-6 max-w-2xl">
      <h1 className="text-2xl font-bold">Settings</h1>

      {/* Account */}
      <Card>
        <CardHeader>
          <CardTitle className="text-sm flex items-center gap-2">
            <User className="h-4 w-4" />
            Account
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-3">
            <span className="text-sm font-medium">{username}</span>
            <Badge variant={isAdmin ? "default" : "secondary"}>
              {isAdmin ? "Admin" : "User"}
            </Badge>
          </div>
        </CardContent>
      </Card>

      {/* API Token */}
      <Card>
        <CardHeader>
          <CardTitle className="text-sm flex items-center gap-2">
            <Terminal className="h-4 w-4" />
            API Token
          </CardTitle>
          <CardDescription>
            Use this header to authenticate API calls from scripts or the command line.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <CopyField
            label="Authorization Header"
            value={authHeader}
            mono
          />
          <CopyField
            label="curl Example"
            value={curlExample}
            mono
          />
          <div className="space-y-1">
            <label className="text-xs font-medium text-muted-foreground">Usage</label>
            <pre className="rounded-md bg-muted p-3 text-xs font-mono overflow-auto leading-relaxed">
{`# List nodes
curl -H "Authorization: ${authHeader}" \\
  ${serverUrl}/node?limit=10

# Upload a file
curl -X POST -H "Authorization: ${authHeader}" \\
  -F "upload=@myfile.fasta" \\
  ${serverUrl}/node

# Download a file
curl -H "Authorization: ${authHeader}" \\
  -o output.dat \\
  "${serverUrl}/node/<node-id>?download"`}
            </pre>
          </div>
        </CardContent>
      </Card>

      {/* Preferences */}
      <Card>
        <CardHeader>
          <CardTitle className="text-sm">Preferences</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium">Theme</p>
              <p className="text-xs text-muted-foreground">Switch between light and dark mode</p>
            </div>
            <Button variant="outline" size="sm" onClick={toggleTheme}>
              {dark ? (
                <><Sun className="mr-2 h-4 w-4" /> Light</>
              ) : (
                <><Moon className="mr-2 h-4 w-4" /> Dark</>
              )}
            </Button>
          </div>

          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium">Default Page Size</p>
              <p className="text-xs text-muted-foreground">Number of nodes per page in the browser</p>
            </div>
            <select
              value={pageSize}
              onChange={(e) => handlePageSizeChange(Number(e.target.value))}
              className="h-9 rounded-md border bg-background px-3 text-sm"
            >
              {[10, 25, 50, 100].map((n) => (
                <option key={n} value={n}>{n}</option>
              ))}
            </select>
          </div>
        </CardContent>
      </Card>

      {/* Server Config (admin only) */}
      {isAdmin && serverInfo && (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm flex items-center gap-2">
              <Shield className="h-4 w-4" />
              Server Configuration
            </CardTitle>
            <CardDescription>Read-only view of the running server configuration.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <div className="grid grid-cols-2 gap-y-2 gap-x-4">
              <span className="text-muted-foreground">Version</span>
              <span className="font-mono">{serverInfo.version}</span>

              <span className="text-muted-foreground">Uptime</span>
              <span>{serverInfo.uptime}</span>

              <span className="text-muted-foreground">URL</span>
              <span className="font-mono text-xs">{serverInfo.url}</span>

              <span className="text-muted-foreground">Contact</span>
              <span>{serverInfo.contact || "—"}</span>

              <span className="text-muted-foreground">Auth Methods</span>
              <span>{serverInfo.auth && serverInfo.auth.length > 0 ? serverInfo.auth.join(", ") : "basic"}</span>

              <span className="text-muted-foreground">Anonymous Read</span>
              <Badge variant={serverInfo.anonymous_permissions.read ? "default" : "outline"}>
                {serverInfo.anonymous_permissions.read ? "allowed" : "denied"}
              </Badge>

              <span className="text-muted-foreground">Anonymous Write</span>
              <Badge variant={serverInfo.anonymous_permissions.write ? "destructive" : "outline"}>
                {serverInfo.anonymous_permissions.write ? "allowed" : "denied"}
              </Badge>

              <span className="text-muted-foreground">Anonymous Delete</span>
              <Badge variant={serverInfo.anonymous_permissions.delete ? "destructive" : "outline"}>
                {serverInfo.anonymous_permissions.delete ? "allowed" : "denied"}
              </Badge>

              <span className="text-muted-foreground">Attribute Indexes</span>
              <span className="font-mono text-xs">
                {serverInfo.attribute_indexes.filter(Boolean).join(", ") || "—"}
              </span>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
