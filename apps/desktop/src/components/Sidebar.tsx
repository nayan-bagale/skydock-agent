import { Activity, Home as HomeIcon } from "lucide-react";

import { useDesktopUi, type DesktopTab } from "../context/DesktopUiContext";
import { cn } from "../utils/cn";
import Button from "./ui/Button";

const navItems: { label: DesktopTab; icon: typeof HomeIcon }[] = [
  { label: "Home", icon: HomeIcon },
  { label: "Sync Activity", icon: Activity },
];

const Sidebar = () => {
  const { activeTab, setActiveTab, menuOpen, setMenuOpen } = useDesktopUi();

  return (
    <aside
      className={cn(
        "relative z-20 flex min-h-screen flex-col bg-sidebar px-3.5 pb-5 pt-[15px]",
        "max-[800px]:fixed max-[800px]:inset-y-auto max-[800px]:bottom-0 max-[800px]:left-0 max-[800px]:top-[60px] max-[800px]:min-h-[calc(100vh-60px)] max-[800px]:w-[245px] max-[800px]:-translate-x-full max-[800px]:shadow-2xl max-[800px]:transition-transform max-[800px]:duration-200",
        menuOpen && "max-[800px]:translate-x-0"
      )}
    >
      <div className="flex h-[43px] items-center gap-[9px] px-2 text-[21px] max-[800px]:hidden">
        {/* <div className="mr-[5px] hidden gap-[7px] min-[1100px]:flex" aria-label="macOS window controls">
          <span className="size-3 rounded-full bg-mac-close" />
          <span className="size-3 rounded-full bg-mac-minimize" />
          <span className="size-3 rounded-full bg-mac-maximize" />
        </div> */}
        <div className="grid size-13 place-items-center">
          <img src="/skydock-logo-resized.png" alt="Skydock Logo" className="w-full h-full object-contain" />
        </div>
        <h1 className="text-2xl ">Skydock</h1>
      </div>

      <nav aria-label="File navigation" className="mt-4 grid gap-1.5 [&_svg]:mr-1.5 [&_svg]:size-5">
        {navItems.map(({ label, icon: Icon }) => (
          <Button
            key={label}
            intent={activeTab === label ? "primary" : "secondary"}
            size="menu"
            onClick={() => {
              setActiveTab(label);
              setMenuOpen(false);
            }}
          >
            <Icon /> <span>{label}</span>
          </Button>
        ))}
      </nav>
    </aside>
  );
};

export default Sidebar;
