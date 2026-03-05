import type { UploadProgress as ProgressData } from "shock-client";
import { formatBytes } from "@/lib/utils";

interface UploadProgressProps {
  progress: ProgressData;
  fileName: string;
}

export function UploadProgress({ progress, fileName }: UploadProgressProps) {
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between text-sm">
        <span className="truncate font-medium">{fileName}</span>
        <span className="text-muted-foreground">{progress.percent}%</span>
      </div>
      <div className="h-2 overflow-hidden rounded-full bg-secondary">
        <div
          className="h-full rounded-full bg-primary transition-all duration-300"
          style={{ width: `${progress.percent}%` }}
        />
      </div>
      <p className="text-xs text-muted-foreground">
        {formatBytes(progress.loaded)} / {formatBytes(progress.total)}
      </p>
    </div>
  );
}
