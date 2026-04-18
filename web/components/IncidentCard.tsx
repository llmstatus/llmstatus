import type { IncidentRef, Severity } from "@/lib/api";

const SEVERITY_STYLE: Record<Severity, { label: string; color: string }> = {
  critical: { label: "Critical", color: "text-[var(--signal-down)]" },
  major:    { label: "Major",    color: "text-[var(--signal-warn)]" },
  minor:    { label: "Minor",    color: "text-[var(--ink-300)]" },
};

function formatDate(iso: string): string {
  return new Date(iso).toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    timeZone: "UTC",
    timeZoneName: "short",
  });
}

export function IncidentCard({ incident }: { incident: IncidentRef }) {
  const { label, color } = SEVERITY_STYLE[incident.severity] ?? SEVERITY_STYLE.minor;

  return (
    <div className="rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-4 py-3">
      <div className="flex items-start justify-between gap-4">
        <p className="text-sm font-medium text-[var(--ink-100)]">{incident.title}</p>
        <span className={`shrink-0 text-xs font-semibold uppercase tracking-wide ${color}`}>
          {label}
        </span>
      </div>
      <p className="mt-1 text-xs text-[var(--ink-400)]">
        Started {formatDate(incident.started_at)}
      </p>
    </div>
  );
}
