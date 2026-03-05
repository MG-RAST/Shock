import { LocationBrowser } from "@/components/admin/LocationBrowser";

export function AdminLocationsPage() {
  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold">Storage Locations</h1>
      <LocationBrowser />
    </div>
  );
}
