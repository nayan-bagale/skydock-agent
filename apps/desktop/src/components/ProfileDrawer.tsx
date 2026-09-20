import { LogOut, UserRound } from "lucide-react";

import StorageTrack from "./ui/StorageTrack";

type ProfileDrawerProps = {
  onClose: () => void;
};

const ProfileDrawer = ({ onClose }: ProfileDrawerProps) => {
  return (
    <>
      <div className="fixed inset-0 z-40" onClick={onClose} />
      <div
        role="menu"
        className="absolute right-4 top-[66px] z-50 flex w-[248px] flex-col items-center rounded-[14px] border border-border bg-card px-[18px] pb-[18px] pt-5 shadow-2xl"
      >
        <div className="mb-2.5 grid size-[62px] place-items-center rounded-full bg-accent text-muted-foreground">
          <UserRound className="size-[30px]" />
        </div>
        <strong className="text-[15px] font-bold">Alex Morgan</strong>
        <span className="mt-0.5 text-xs text-muted-foreground">alex.morgan@skydock.io</span>
        <button
          className="mt-3.5 inline-flex w-full items-center gap-2 rounded-[9px] border border-destructive/35 px-3 py-[9px] text-[13px] font-semibold text-destructive hover:bg-destructive/10"
          onClick={onClose}
        >
          <LogOut className="size-[15px]" /> Disconnect Account
        </button>
        <div className="mt-3.5 w-full rounded-[10px] border border-border bg-sidebar/60 px-3.5 pb-3.5 pt-[13px]">
          <div className="flex items-baseline justify-between text-xs text-muted-foreground">
            <span>Storage</span>
            <strong className="text-[13px] font-bold text-foreground">45.2 GB / 100 GB</strong>
          </div>
          <StorageTrack className="mt-2.5" />
          <button
            className="mt-3 w-full rounded-[9px] bg-primary px-3 py-[9px] text-[13px] font-semibold text-primary-foreground hover:bg-primary/90"
            onClick={onClose}
          >
            Upgrade
          </button>
        </div>
      </div>
    </>
  );
};

export default ProfileDrawer;
