import { CheckCircle2, Loader2, WifiOff } from "lucide-react";

import { useAuth } from "../context/AuthContext";
import { useAgentConnection } from "../hooks/useAgentConnection";
import { changeBytes, getStorageUsage } from "../utils/changeBytes";
import Button from "./ui/Button";
import StorageTrack from "./ui/StorageTrack";

const HomeOverview = () => {
  const { connected, status, attempt, agentVersion, retry } = useAgentConnection();
  const { user } = useAuth();
  const storageLimit = user?.plan.storageLimit ?? 0;
  const usedStorage = user?.usedStorage ?? 0;
  const { usedPercentage } = getStorageUsage(storageLimit, usedStorage);
  const connecting = status === "connecting";

  return (
    <section className="mt-[34px] grid grid-cols-2 gap-3.5 max-[800px]:grid-cols-1">
      <div className="rounded-xl border border-border bg-card px-6 py-[22px] shadow-sm">
        <h2 className="mb-3.5 text-base font-semibold">Sync Status</h2>
        <div
          className={`inline-flex items-center gap-[9px] rounded-full px-3.5 py-[9px] text-sm font-semibold ${
            connected ? "bg-success-bg text-success" : "bg-muted text-muted-foreground"
          }`}
        >
          {connecting ? (
            <Loader2 className="w-[17px] animate-spin" />
          ) : connected ? (
            <CheckCircle2 className="w-[17px]" />
          ) : (
            <WifiOff className="w-[17px]" />
          )}
          {connecting ? "Connecting…" : connected ? "Agent connected" : "Agent disconnected"}
        </div>
        <p className="mt-3.5 text-[13px] text-muted-foreground">
          {connected
            ? `Sync agent online${agentVersion ? ` · v${agentVersion}` : ""}`
            : connecting
              ? `Trying to reach the sync agent${attempt > 0 ? ` · attempt ${attempt} of 5` : ""}`
              : "Start the Go agent and ensure ZMQ is listening on ipc:///tmp/skydock-agent.sock"}
        </p>
        {status === "failed" && (
          <Button intent="primary" size="md" className="mt-3.5" onClick={() => void retry()}>
            Try again
          </Button>
        )}
      </div>
      <div className="rounded-xl border border-border bg-card px-6 py-[22px] shadow-sm">
        <h2 className="mb-3.5 text-base font-semibold">Storage</h2>
        <strong className="block text-[26px] font-bold">
          {user ? changeBytes(usedStorage) : "—"}{" "}
          <span className="text-sm font-normal text-muted-foreground">
            / {user ? changeBytes(storageLimit) : "—"}
          </span>
        </strong>
        <StorageTrack usedPercentage={usedPercentage} />
        <p className="mt-3.5 text-[13px] text-muted-foreground">
          {user ? `${usedPercentage}% of your plan used` : "Loading storage…"}
        </p>
      </div>
    </section>
  );
};

export default HomeOverview;
