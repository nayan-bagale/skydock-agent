import { cn } from "../../utils/cn";

function StorageTrack({ className }: { className?: string }) {
  return (
    <div className={cn("mt-3.5 h-1.5 overflow-hidden rounded-full bg-muted", className)}>
      <span className="block h-full w-[45.2%] rounded-full bg-primary" />
    </div>
  );
}

export default StorageTrack;
