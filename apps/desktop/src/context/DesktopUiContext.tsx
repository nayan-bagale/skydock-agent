import { createContext, useContext, useMemo, useState, type ReactNode } from "react";

export type DesktopTab = "Home" | "Sync Activity";

type DesktopUiContextValue = {
  activeTab: DesktopTab;
  setActiveTab: (tab: DesktopTab) => void;
  menuOpen: boolean;
  setMenuOpen: (open: boolean | ((value: boolean) => boolean)) => void;
};

const DesktopUiContext = createContext<DesktopUiContextValue | null>(null);

export function DesktopUiProvider({ children }: { children: ReactNode }) {
  const [activeTab, setActiveTab] = useState<DesktopTab>("Home");
  const [menuOpen, setMenuOpen] = useState(false);

  const value = useMemo(
    () => ({ activeTab, setActiveTab, menuOpen, setMenuOpen }),
    [activeTab, menuOpen]
  );

  return <DesktopUiContext.Provider value={value}>{children}</DesktopUiContext.Provider>;
}

export function useDesktopUi() {
  const context = useContext(DesktopUiContext);
  if (!context) {
    throw new Error("useDesktopUi must be used within a DesktopUiProvider");
  }
  return context;
}
