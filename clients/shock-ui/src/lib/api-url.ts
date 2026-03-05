/** Returns the Shock API base URL.
 *  - Dev mode: direct to :7445 (bypasses Vite proxy, avoids root-path conflict)
 *  - Production (embedded): same-origin (empty string)
 */
export function getApiUrl(): string {
  if (import.meta.env.DEV) {
    return `${window.location.protocol}//${window.location.hostname}:7445`;
  }
  return "";
}
