import { useAuth } from "./use-auth";

export function useIsAdmin(): boolean {
  const { isAdmin } = useAuth();
  return isAdmin;
}
