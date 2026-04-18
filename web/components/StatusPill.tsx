import type { ProviderStatus } from "@/lib/api";

// Brand spec §6.1: dot-only, no background pill, all-caps text, color inherits from status.
const CONFIG: Record<ProviderStatus, { label: string; dot: string; text: string }> = {
  operational: {
    label: "Operational",
    dot: "bg-[var(--signal-ok)]",
    text: "text-[var(--signal-ok)]",
  },
  degraded: {
    label: "Degraded",
    dot: "bg-[var(--signal-warn)]",
    text: "text-[var(--signal-warn)]",
  },
  down: {
    label: "Down",
    dot: "bg-[var(--signal-down)]",
    text: "text-[var(--signal-down)]",
  },
};

export function StatusPill({ status }: { status: ProviderStatus }) {
  const { label, dot, text } = CONFIG[status];
  return (
    <span className={`inline-flex items-center gap-2 ${text}`}>
      <span className={`h-2 w-2 rounded-full ${dot}`} aria-hidden="true" />
      <span className="text-[11px] font-semibold uppercase tracking-[0.05em]">{label}</span>
    </span>
  );
}
