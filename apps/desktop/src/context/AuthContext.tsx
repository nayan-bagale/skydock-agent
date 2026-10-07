import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";

import {
  EventAuthChanged,
  EventAuthLogout,
  EventAuthSession,
  EventAuthStart,
} from "../../electron/agent-protocol";

export type AuthUser = {
  name: string;
  email: string;
  profileUrl: string;
  usedStorage: number;
  plan: { storageLimit: number };
};

type AuthSessionAck = {
  ok?: boolean;
  authenticated?: boolean;
  accessToken?: string;
  apiBaseUrl?: string;
  url?: string;
  error?: string;
};

type AuthChangedPayload = {
  authenticated?: boolean;
  accessToken?: string;
  apiBaseUrl?: string;
};

type AuthContextValue = {
  user: AuthUser | null;
  accessToken: string;
  apiBaseUrl: string;
  isAuthenticated: boolean;
  isLoading: boolean;
  setUser: (user: AuthUser | null) => void;
  login: () => Promise<void>;
  logout: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

function applyAuthState(
  authenticated: boolean,
  accessToken: string | undefined,
  apiBaseUrl: string | undefined,
  setters: {
    setIsAuthenticated: (v: boolean) => void;
    setAccessToken: (v: string) => void;
    setApiBaseUrl: (v: string) => void;
    setUser: (v: AuthUser | null) => void;
  },
) {
  setters.setIsAuthenticated(authenticated);
  if (!authenticated || !accessToken) {
    setters.setAccessToken("");
    setters.setUser(null);
    return;
  }
  setters.setAccessToken(accessToken);
  if (apiBaseUrl) {
    setters.setApiBaseUrl(apiBaseUrl);
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [accessToken, setAccessToken] = useState("");
  const [apiBaseUrl, setApiBaseUrl] = useState("");
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;

    const loadSession = async () => {
      try {
        const ack = (await window.electron.zmq.emit(
          EventAuthSession,
          {},
          true,
        )) as AuthSessionAck;
        if (!cancelled && ack?.ok && typeof ack.authenticated === "boolean") {
          applyAuthState(ack.authenticated, ack.accessToken, ack.apiBaseUrl, {
            setIsAuthenticated,
            setAccessToken,
            setApiBaseUrl,
            setUser,
          });
        }
      } catch {
        if (!cancelled) {
          setIsAuthenticated(false);
          setAccessToken("");
          setUser(null);
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false);
        }
      }
    };

    void loadSession();

    const offChanged = window.electron.zmq.on(EventAuthChanged, (envelope) => {
      const data = envelope.data as AuthChangedPayload | undefined;
      if (typeof data?.authenticated !== "boolean") {
        return;
      }
      applyAuthState(data.authenticated, data.accessToken, data.apiBaseUrl, {
        setIsAuthenticated,
        setAccessToken,
        setApiBaseUrl,
        setUser,
      });
    });

    return () => {
      cancelled = true;
      offChanged();
    };
  }, []);

  const login = useCallback(async () => {
    const ack = (await window.electron.zmq.emit(
      EventAuthStart,
      {},
      true,
    )) as AuthSessionAck;
    if (!ack?.ok || !ack.url) {
      throw new Error(ack?.error ?? "Could not start sign-in");
    }
    await window.electron.openExternal(ack.url);
  }, []);

  const logout = useCallback(async () => {
    await window.electron.zmq.emit(EventAuthLogout, {}, true);
    setIsAuthenticated(false);
    setAccessToken("");
    setUser(null);
  }, []);

  const value = useMemo(
    () => ({
      user,
      accessToken,
      apiBaseUrl,
      isAuthenticated,
      isLoading,
      setUser,
      login,
      logout,
    }),
    [
      user,
      accessToken,
      apiBaseUrl,
      isAuthenticated,
      isLoading,
      login,
      logout,
    ],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
