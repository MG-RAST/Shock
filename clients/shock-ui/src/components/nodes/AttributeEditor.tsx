import { useState } from "react";
import { useUpdateAttributes } from "shock-client/react";
import { Button } from "@/components/ui/button";
import { Save, RotateCcw } from "lucide-react";

interface AttributeEditorProps {
  nodeId: string;
  attributes: unknown;
}

export function AttributeEditor({ nodeId, attributes }: AttributeEditorProps) {
  const original = JSON.stringify(attributes, null, 2) || "{}";
  const [text, setText] = useState(original);
  const [parseError, setParseError] = useState<string | null>(null);
  const updateAttrs = useUpdateAttributes(nodeId);

  const isDirty = text !== original;

  const handleSave = () => {
    try {
      const parsed = JSON.parse(text);
      setParseError(null);
      updateAttrs.mutate({ attributes: parsed });
    } catch (e) {
      setParseError(e instanceof Error ? e.message : "Invalid JSON");
    }
  };

  const handleReset = () => {
    setText(original);
    setParseError(null);
  };

  return (
    <div className="space-y-3">
      <textarea
        value={text}
        onChange={(e) => {
          setText(e.target.value);
          setParseError(null);
        }}
        className="h-64 w-full rounded-md border bg-muted/30 p-3 font-mono text-xs focus:outline-none focus:ring-1 focus:ring-ring"
        spellCheck={false}
      />
      {parseError && (
        <p className="text-sm text-destructive">{parseError}</p>
      )}
      {updateAttrs.error && (
        <p className="text-sm text-destructive">{updateAttrs.error.message}</p>
      )}
      <div className="flex gap-2">
        <Button
          size="sm"
          onClick={handleSave}
          disabled={!isDirty || updateAttrs.isPending}
        >
          <Save className="mr-1 h-3 w-3" />
          {updateAttrs.isPending ? "Saving..." : "Save"}
        </Button>
        <Button variant="outline" size="sm" onClick={handleReset} disabled={!isDirty}>
          <RotateCcw className="mr-1 h-3 w-3" />
          Reset
        </Button>
      </div>
    </div>
  );
}
