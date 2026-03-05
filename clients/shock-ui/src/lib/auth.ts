const TOKEN_KEY = "shock_auth_token";
const USER_KEY = "shock_auth_user";
const ADMIN_KEY = "shock_is_admin";

export function getStoredToken(): string | undefined {
  return sessionStorage.getItem(TOKEN_KEY) ?? undefined;
}

export function getStoredUser(): string | undefined {
  return sessionStorage.getItem(USER_KEY) ?? undefined;
}

export function getStoredAdmin(): boolean {
  return sessionStorage.getItem(ADMIN_KEY) === "true";
}

export function storeAuth(token: string, username: string, isAdmin: boolean): void {
  sessionStorage.setItem(TOKEN_KEY, token);
  sessionStorage.setItem(USER_KEY, username);
  sessionStorage.setItem(ADMIN_KEY, String(isAdmin));
}

export function clearAuth(): void {
  sessionStorage.removeItem(TOKEN_KEY);
  sessionStorage.removeItem(USER_KEY);
  sessionStorage.removeItem(ADMIN_KEY);
}
