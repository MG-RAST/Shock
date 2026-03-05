import { TraceControls } from "@/components/admin/TraceControls";

export function AdminTracePage() {
  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold">Execution Trace</h1>
      <TraceControls />
    </div>
  );
}
