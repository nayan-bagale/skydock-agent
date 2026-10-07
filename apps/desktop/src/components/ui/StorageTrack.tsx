import { cn } from "../../utils/cn";

type StorageTrackProps = {
  className?: string;
  usedPercentage?: number;
};

function StorageTrack({ className, usedPercentage = 0 }: StorageTrackProps) {
  const width = `${Math.min(100, usedPercentage)}%`;
  return (
    <div className={cn("mt-3.5 h-1.5 overflow-hidden rounded-full bg-muted", className)}>
      <span
        className="block h-full rounded-full bg-primary"
        style={{ width }}
      />
    </div>
  );
}

export default StorageTrack;
