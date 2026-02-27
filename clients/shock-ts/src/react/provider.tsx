import {
  createContext,
  useContext,
  useMemo,
  useRef,
  type ReactNode,
} from "react";
import {
  QueryClient,
  QueryClientProvider,
} from "@tanstack/react-query";
import { ShockClient } from "../client.js";

const ShockContext = createContext<ShockClient | null>(null);

export interface ShockProviderProps {
  /** Base URL of the Shock server. */
  url: string;
  /** Static auth token. */
  token?: string;
  /** Dynamic token getter (takes priority over `token`). */
  getToken?: () => string | undefined;
  /** Optional externally-managed QueryClient. */
  queryClient?: QueryClient;
  children: ReactNode;
}

/**
 * Provides a `ShockClient` to all descendant hooks.
 *
 * When `token` changes, calls `client.setToken()` rather than creating
 * a new client instance — this avoids invalidating the entire query cache.
 */
export function ShockProvider({
  url,
  token,
  getToken,
  queryClient: externalQc,
  children,
}: ShockProviderProps) {
  // Create an internal QueryClient only if the caller didn't provide one
  const internalQcRef = useRef<QueryClient>();
  if (!externalQc && !internalQcRef.current) {
    internalQcRef.current = new QueryClient();
  }
  const qc = externalQc ?? internalQcRef.current!;

  // Memoize the client on [url, getToken]. Token changes go through setToken.
  const client = useMemo(
    () => new ShockClient({ url, token, getToken }),
    [url, getToken] // eslint-disable-line react-hooks/exhaustive-deps
  );

  // Update token without re-creating client
  useMemo(() => {
    if (!getToken) {
      client.setToken(token);
    }
  }, [client, token, getToken]);

  return (
    <QueryClientProvider client={qc}>
      <ShockContext.Provider value={client}>{children}</ShockContext.Provider>
    </QueryClientProvider>
  );
}

/** Returns the ShockClient from the nearest ShockProvider. */
export function useShockClient(): ShockClient {
  const client = useContext(ShockContext);
  if (!client) {
    throw new Error("useShockClient must be used within a <ShockProvider>");
  }
  return client;
}
