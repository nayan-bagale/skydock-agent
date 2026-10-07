import { useEffect, useRef, useState } from "react";

import { useAuth } from "../context/AuthContext";

type UserInfoResponse = {
  name?: string;
  email?: string;
  picture?: string;
  usedStorage?: number;
  plan?: { storageLimit?: number };
};

/** Loads `/auth/user-info` once per sign-in; does not refetch on access-token refresh. */
export function useGetUserInfo() {
  const { user, accessToken, apiBaseUrl, isAuthenticated, setUser } =
    useAuth();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const fetchedForSessionRef = useRef(false);
  const accessTokenRef = useRef(accessToken);
  accessTokenRef.current = accessToken;

  useEffect(() => {
    if (!isAuthenticated) {
      fetchedForSessionRef.current = false;
      return;
    }
    if (!accessToken || !apiBaseUrl || fetchedForSessionRef.current) {
      return;
    }

    fetchedForSessionRef.current = true;
    let cancelled = false;
    setIsLoading(true);
    setError(null);

    const token = accessTokenRef.current;
    void fetch(`${apiBaseUrl}/auth/user-info`, {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then(async (res) => {
        if (!res.ok) {
          throw new Error("Failed to load user");
        }
        return (await res.json()) as UserInfoResponse;
      })
      .then((data) => {
        if (cancelled) {
          return;
        }
        setUser({
          name: data.name ?? "",
          email: data.email ?? "",
          profileUrl: data.picture ?? "",
          usedStorage: data.usedStorage ?? 0,
          plan: { storageLimit: data.plan?.storageLimit ?? 0 },
        });
      })
      .catch((err: unknown) => {
        if (cancelled) {
          return;
        }
        fetchedForSessionRef.current = false;
        setError(
          err instanceof Error ? err.message : "Failed to load user",
        );
      })
      .finally(() => {
        if (!cancelled) {
          setIsLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [isAuthenticated, accessToken, apiBaseUrl, setUser]);

  return { user, isLoading, error };
}
