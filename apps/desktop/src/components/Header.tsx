import { useState } from "react";
import { Maximize2, Menu, Minus, UserRound, X } from "lucide-react";

import { useDesktopUi } from "../context/DesktopUiContext";
import ProfileDrawer from "./ProfileDrawer";
import Button from "./ui/Button";
import { useOS } from "../hooks/useOS";

const Header = () => {
  const { setMenuOpen } = useDesktopUi();
  const [profileOpen, setProfileOpen] = useState(false);
  const { isWindows } = useOS()
  return (
    <header className="relative flex h-[60px] items-center pr-4 bg-card/60">
      <Button
        className="ml-2 hidden max-[800px]:inline-flex"
        intent="secondary"
        size="icon"
        onClick={() => setMenuOpen((value) => !value)}
        aria-label="Toggle navigation"
      >
        <Menu />
      </Button>
      <div className="ml-auto flex items-center gap-[5px] max-[540px]:hidden">
        <Button
          intent="secondary"
          size="icon"
          aria-label="User account"
          aria-expanded={profileOpen}
          onClick={() => setProfileOpen((value) => !value)}
        >
          <UserRound />
        </Button>
      </div>
      {profileOpen && <ProfileDrawer onClose={() => setProfileOpen(false)} />}
      {isWindows && <div className="ml-[18px] flex gap-2 self-stretch border-l border-border max-[800px]:hidden [&_svg]:w-[15px]" aria-label="Windows controls">
        <Button intent="window" size="window" aria-label="Minimize">
          <Minus />
        </Button>
        <Button intent="window" size="window" aria-label="Maximize">
          <Maximize2 />
        </Button>
        <Button intent="window" size="window" aria-label="Close">
          <X />
        </Button>
      </div>}
    </header>
  );
};

export default Header;
