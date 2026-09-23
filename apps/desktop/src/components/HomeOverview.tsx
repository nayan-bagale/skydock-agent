import { CheckCircle2, WifiOff } from "lucide-react";

import { useAgentConnection } from "../hooks/useAgentConnection";
import StorageTrack from "./ui/StorageTrack";

const HomeOverview = () => {
  const { connected, agentVersion } = useAgentConnection();

  return (
    <section className="mt-[34px] grid grid-cols-2 gap-3.5 max-[800px]:grid-cols-1">
      <div className="rounded-xl border border-border bg-card px-6 py-[22px] shadow-sm">
        <h2 className="mb-3.5 text-base font-semibold">Sync Status</h2>
        <div
          className={`inline-flex items-center gap-[9px] rounded-full px-3.5 py-[9px] text-sm font-semibold ${
            connected ? "bg-success-bg text-success" : "bg-muted text-muted-foreground"
          }`}
        >
          {connected ? <CheckCircle2 className="w-[17px]" /> : <WifiOff className="w-[17px]" />}
          {connected ? "Agent connected" : "Agent disconnected"}
        </div>
        <p className="mt-3.5 text-[13px] text-muted-foreground">
          {connected
            ? `Sync agent online${agentVersion ? ` · v${agentVersion}` : ""}`
            : "Start the Go agent and ensure ZMQ is listening on port 17300"}
        </p>
      </div>
      <div className="rounded-xl border border-border bg-card px-6 py-[22px] shadow-sm">
        <h2 className="mb-3.5 text-base font-semibold">Storage</h2>
        <strong className="block text-[26px] font-bold">
          45.2 GB <span className="text-sm font-normal text-muted-foreground">/ 100 GB</span>
        </strong>
        <StorageTrack />
        <p className="mt-3.5 text-[13px] text-muted-foreground">45.2% of your plan used</p>
      </div>
    </section>
  );
};

export default HomeOverview;
