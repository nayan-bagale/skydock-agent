import { CheckCircle2 } from "lucide-react";

import { useDesktopUi } from "../context/DesktopUiContext";
import HomeOverview from "./HomeOverview";
import SyncActivity from "./SyncActivity";

const MainContent = () => {
  const { activeTab } = useDesktopUi();

  return (
    <div className="px-6 pb-10 pt-[23px] max-[800px]:px-4 max-[800px]:py-5">
      <div className="flex items-center justify-between gap-4 max-[540px]:flex-wrap">
        <div>
          <h1 className="text-[31px] font-bold leading-[1.15] max-[540px]:text-[26px]">{activeTab}</h1>
          <p className="mt-1.5 text-sm text-muted-foreground">
            {activeTab === "Home" ? "Your sync and storage at a glance." : "Recent synchronization events."}
          </p>
        </div>
        {activeTab === "Sync Activity" && (
          <span className="inline-flex items-center gap-2 whitespace-nowrap rounded-full border border-border bg-card px-3.5 py-2 text-[13px] font-semibold text-muted-foreground max-[540px]:px-3 max-[540px]:py-1.5 max-[540px]:text-xs">
            <CheckCircle2 className="w-[15px] text-success" /> Last synced just now
          </span>
        )}
      </div>

      {activeTab === "Home" && <HomeOverview />}
      {activeTab === "Sync Activity" && <SyncActivity />}
    </div>
  );
};

export default MainContent;
