import { CheckCircle2, Zap } from "lucide-react";

import { cn } from "../utils/cn";

const syncEvents = [
  { name: "Hero_Banner_v2.png", size: "2.4 MB", status: "Syncing" as const },
  { name: "Q4 Strategy Deck.pptx", size: "18.7 MB", status: "Complete" as const },
  { name: "Contract_AcmeCorp_Signed.pdf", size: "1.1 MB", status: "Complete" as const },
  { name: "Engineering Specs", size: "326 MB", status: "Complete" as const },
  { name: "Marketing Assets 2024", size: "2.1 GB", status: "Complete" as const },
];

const SyncActivity = () => {
  return (
    <section aria-label="Sync activity" className="mt-[34px] grid gap-2">
      {syncEvents.map((event) => (
        <div
          key={event.name}
          className="grid min-h-[58px] grid-cols-[minmax(240px,1fr)_110px_128px] items-center rounded-xl border border-border/70 bg-card px-[17px] text-xs text-muted-foreground shadow-sm max-[800px]:grid-cols-[minmax(140px,1fr)_118px]"
        >
          <span className="flex min-w-0 items-center gap-3 text-sm text-foreground">
            <span className="truncate">{event.name}</span>
          </span>
          <span className="truncate max-[800px]:hidden">{event.size}</span>
          <span
            className={cn(
              "flex items-center gap-1.5 justify-self-end font-semibold [&_svg]:size-3.5 [&_svg]:shrink-0",
              event.status === "Syncing" ? "text-warning" : "text-success"
            )}
          >
            {event.status === "Syncing" ? <Zap /> : <CheckCircle2 />}
            <span>{event.status}</span>
          </span>
        </div>
      ))}
    </section>
  );
};

export default SyncActivity;
