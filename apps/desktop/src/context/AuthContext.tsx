import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from "react";

export type AuthUser = {
  name: string;
  email: string;
  profileUrl: string;
};

type AuthContextValue = {
  user: AuthUser | null;
  isAuthenticated: boolean;
  login: () => void;
  logout: () => void;
};

const AuthContext = createContext<AuthContextValue | null>(null);

const DEMO_USER: AuthUser = {
  name: "Alex Morgan",
  email: "alex.morgan@skydock.io",
  profileUrl: "https://i.pravatar.cc/150?u=alex.morgan@skydock.io",
};

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);

  const login = useCallback(() => {
    setUser(DEMO_USER);
  }, []);

  const logout = useCallback(() => {
    setUser(null);
  }, []);

  const value = useMemo(
    () => ({
      user,
      isAuthenticated: user !== null,
      login,
      logout,
    }),
    [user, login, logout]
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
