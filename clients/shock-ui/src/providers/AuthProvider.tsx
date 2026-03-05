import { useState, useCallback, type ReactNode } from "react";
import { AuthContext, type AuthState } from "../hooks/use-auth";
import {
  getStoredToken,
  getStoredUser,
  getStoredAdmin,
  storeAuth,
  clearAuth,
} from "../lib/auth";
import { encodeBasicAuth } from "../lib/utils";
import { getApiUrl } from "../lib/api-url";

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | undefined>(getStoredToken);
  const [username, setUsername] = useState<string | undefined>(getStoredUser);
  const [isAdmin, setIsAdmin] = useState(getStoredAdmin);

  const login = useCallback(async (user: string, password: string) => {
    const encoded = encodeBasicAuth(user, password);
    const baseUrl = getApiUrl();

    // Validate credentials by fetching node list
    const res = await fetch(`${baseUrl}/node?limit=1`, {
      headers: { Authorization: `Basic ${encoded}` },
    });

    if (!res.ok) {
      const body = await res.json().catch(() => null);
      const msg = body?.error?.[0] ?? `Authentication failed (HTTP ${res.status})`;
      throw new Error(msg);
    }

    // Check admin status by trying /locker
    let admin = false;
    try {
      const adminRes = await fetch(`${baseUrl}/locker`, {
        headers: { Authorization: `Basic ${encoded}` },
      });
      admin = adminRes.ok;
    } catch {
      // Not admin, that's fine
    }

    storeAuth(encoded, user, admin);
    setToken(encoded);
    setUsername(user);
    setIsAdmin(admin);
  }, []);

  const logout = useCallback(() => {
    clearAuth();
    setToken(undefined);
    setUsername(undefined);
    setIsAdmin(false);
  }, []);

  const value: AuthState = {
    token,
    username,
    isAdmin,
    isAuthenticated: Boolean(token),
    login,
    logout,
  };

  return (
    <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
  );
}
