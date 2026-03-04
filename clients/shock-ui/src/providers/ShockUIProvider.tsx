import { type ReactNode, useMemo } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ShockProvider } from "shock-client/react";
import { AuthProvider } from "./AuthProvider";
import { useAuth } from "../hooks/use-auth";
import { getApiUrl } from "../lib/api-url";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

function ShockClientBridge({ children }: { children: ReactNode }) {
  const { token } = useAuth();

  const getToken = useMemo(() => {
    if (!token) return undefined;
    const t = token;
    return () => t;
  }, [token]);

  return (
    <ShockProvider
      url={getApiUrl()}
      getToken={getToken}
      authType="basic"
      queryClient={queryClient}
    >
      {children}
    </ShockProvider>
  );
}

export function ShockUIProvider({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <ShockClientBridge>{children}</ShockClientBridge>
      </AuthProvider>
    </QueryClientProvider>
  );
}
