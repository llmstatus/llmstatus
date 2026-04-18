import type { ProviderStatus } from "@/lib/api";

const CONFIG: Record<ProviderStatus, { label: string; dot: string; bg: string; text: string }> = {
  operational: {
    label: "Operational",
    dot: "bg-[var(--signal-ok)]",
    bg: "bg-[var(--signal-ok-bg)]",
    text: "text-[var(--signal-ok)]",
  },
  degraded: {
    label: "Degraded",
    dot: "bg-[var(--signal-warn)]",
    bg: "bg-[var(--signal-warn-bg)]",
    text: "text-[var(--signal-warn)]",
  },
  down: {
    label: "Down",
    dot: "bg-[var(--signal-down)]",
    bg: "bg-[var(--signal-down-bg)]",
    text: "text-[var(--signal-down)]",
  },
};

export function StatusPill({ status }: { status: ProviderStatus }) {
  const { label, dot, bg, text } = CONFIG[status];
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium ${bg} ${text}`}
    >
      <span className={`h-1.5 w-1.5 rounded-full ${dot}`} aria-hidden="true" />
      {label}
    </span>
  );
}
