import { useNodeAcl, useAddAcl, useRemoveAcl } from "shock-client/react";
import type { AclType } from "shock-client";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Plus, X, Globe } from "lucide-react";

interface AclPanelProps {
  nodeId: string;
}

function AclUserList({
  nodeId,
  type,
  users,
}: {
  nodeId: string;
  type: AclType;
  users: string[];
}) {
  const addAcl = useAddAcl(nodeId);
  const removeAcl = useRemoveAcl(nodeId);
  const [newUser, setNewUser] = useState("");

  const handleAdd = () => {
    const trimmed = newUser.trim();
    if (!trimmed) return;
    addAcl.mutate({ type, users: [trimmed] });
    setNewUser("");
  };

  return (
    <div className="space-y-2">
      <div className="flex flex-wrap gap-1">
        {users.length === 0 && (
          <span className="text-xs text-muted-foreground">No users</span>
        )}
        {users.map((user) => (
          <Badge key={user} variant="secondary" className="gap-1">
            {user}
            <button
              onClick={() => removeAcl.mutate({ type, users: [user] })}
              className="ml-1 hover:text-destructive"
            >
              <X className="h-3 w-3" />
            </button>
          </Badge>
        ))}
      </div>
      <div className="flex gap-2">
        <Input
          value={newUser}
          onChange={(e) => setNewUser(e.target.value)}
          placeholder="Add user..."
          className="h-8 text-xs"
          onKeyDown={(e) => e.key === "Enter" && handleAdd()}
        />
        <Button variant="outline" size="sm" onClick={handleAdd} disabled={!newUser.trim()}>
          <Plus className="h-3 w-3" />
        </Button>
      </div>
    </div>
  );
}

export function AclPanel({ nodeId }: AclPanelProps) {
  const { data: acl, isLoading } = useNodeAcl(nodeId);
  const addAcl = useAddAcl(nodeId);
  const removeAcl = useRemoveAcl(nodeId);

  if (isLoading) {
    return <div className="space-y-3"><Skeleton className="h-24" /><Skeleton className="h-24" /></div>;
  }

  if (!acl) return null;

  const togglePublic = (type: "public_read" | "public_write" | "public_delete", current: boolean) => {
    if (current) {
      removeAcl.mutate({ type, users: [] });
    } else {
      addAcl.mutate({ type, users: [] });
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h4 className="mb-1 text-sm font-medium">Owner</h4>
        <Badge>{acl.owner}</Badge>
      </div>

      <div>
        <h4 className="mb-1 text-sm font-medium flex items-center gap-2">
          Public Access
          <Globe className="h-3 w-3 text-muted-foreground" />
        </h4>
        <div className="flex gap-2">
          {(["read", "write", "delete"] as const).map((perm) => (
            <Button
              key={perm}
              variant={acl.public[perm] ? "default" : "outline"}
              size="sm"
              onClick={() => togglePublic(`public_${perm}`, acl.public[perm])}
            >
              {perm}
            </Button>
          ))}
        </div>
      </div>

      {(["read", "write", "delete"] as const).map((type) => (
        <div key={type}>
          <h4 className="mb-2 text-sm font-medium capitalize">{type}</h4>
          <AclUserList nodeId={nodeId} type={type} users={acl[type]} />
        </div>
      ))}
    </div>
  );
}
