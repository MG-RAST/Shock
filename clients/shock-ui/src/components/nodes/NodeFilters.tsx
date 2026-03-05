import { useState, type FormEvent } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Search, X } from "lucide-react";

interface NodeFiltersProps {
  onSearch: (query: Record<string, string>) => void;
  onSort: (field: string, direction: "asc" | "desc") => void;
  onClear: () => void;
}

export function NodeFilters({ onSearch, onSort, onClear }: NodeFiltersProps) {
  const [filterText, setFilterText] = useState("");
  const [sortField, setSortField] = useState("created_on");
  const [sortDir, setSortDir] = useState<"asc" | "desc">("desc");

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    const query: Record<string, string> = {};
    // Parse key=value pairs separated by spaces
    const parts = filterText.trim().split(/\s+/);
    for (const part of parts) {
      const eqIdx = part.indexOf("=");
      if (eqIdx > 0) {
        query[part.slice(0, eqIdx)] = part.slice(eqIdx + 1);
      }
    }
    onSearch(query);
  };

  const handleSortChange = (field: string) => {
    const newDir = field === sortField && sortDir === "desc" ? "asc" : "desc";
    setSortField(field);
    setSortDir(newDir);
    onSort(field, newDir);
  };

  return (
    <div className="space-y-3">
      <form onSubmit={handleSubmit} className="flex gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={filterText}
            onChange={(e) => setFilterText(e.target.value)}
            placeholder='Filter: key=value (e.g. attributes.type=metagenome)'
            className="pl-9"
          />
        </div>
        <Button type="submit" variant="secondary" size="default">
          Search
        </Button>
        {filterText && (
          <Button
            type="button"
            variant="ghost"
            size="icon"
            onClick={() => {
              setFilterText("");
              onClear();
            }}
          >
            <X className="h-4 w-4" />
          </Button>
        )}
      </form>
      <div className="flex items-center gap-2 text-xs text-muted-foreground">
        <span>Sort by:</span>
        {["created_on", "file.name", "file.size", "last_modified"].map((field) => (
          <button
            key={field}
            onClick={() => handleSortChange(field)}
            className={`rounded px-2 py-1 transition-colors ${
              sortField === field ? "bg-accent text-accent-foreground" : "hover:bg-accent"
            }`}
          >
            {field.replace("file.", "").replace("_", " ")}
            {sortField === field && (sortDir === "asc" ? " ↑" : " ↓")}
          </button>
        ))}
      </div>
    </div>
  );
}
