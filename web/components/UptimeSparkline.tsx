import type { HistoryBucket } from "@/lib/api";

interface Props {
  buckets: HistoryBucket[];
  /** Number of days to show; padded with empty slots when data is sparse. */
  days?: number;
}

function uptimeColor(uptime: number, hasData: boolean): string {
  if (!hasData) return "bg-[var(--ink-600)]";
  if (uptime >= 0.99) return "bg-[var(--signal-ok)]";
  if (uptime >= 0.95) return "bg-[var(--signal-warn)]";
  return "bg-[var(--signal-down)]";
}

function uptimeHeight(uptime: number, hasData: boolean): string {
  if (!hasData) return "h-2"; // gray stub for missing data
  // Map 0–1 uptime to 20–100% of container height (always at least a sliver)
  const pct = Math.max(20, Math.round(uptime * 100));
  // Tailwind can't interpolate arbitrary values at runtime, so use inline style.
  return `h-[${pct}%]`;
}

function formatDay(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    timeZone: "UTC",
  });
}

export function UptimeSparkline({ buckets, days = 30 }: Props) {
  // Build a day-keyed map for O(1) lookup.
  const byDay = new Map<string, HistoryBucket>();
  for (const b of buckets) {
    const key = b.timestamp.slice(0, 10); // "YYYY-MM-DD"
    byDay.set(key, b);
  }

  // Generate the last `days` day keys in ascending order.
  const slots: string[] = [];
  const now = new Date();
  for (let i = days - 1; i >= 0; i--) {
    const d = new Date(now);
    d.setUTCDate(d.getUTCDate() - i);
    slots.push(d.toISOString().slice(0, 10));
  }

  const overall =
    buckets.length === 0
      ? null
      : buckets.reduce((sum, b) => sum + b.uptime, 0) / buckets.length;

  return (
    <div>
      {/* Summary line */}
      <div className="mb-2 flex items-center justify-between text-xs text-[var(--ink-400)]">
        <span>30-day uptime</span>
        {overall !== null && (
          <span className="font-medium text-[var(--ink-200)]">
            {(overall * 100).toFixed(2)} %
          </span>
        )}
      </div>

      {/* Bars */}
      <div className="flex h-10 items-end gap-px" aria-label="30-day uptime chart">
        {slots.map((day) => {
          const b = byDay.get(day);
          const hasData = b !== undefined && b.total > 0;
          const uptime = b?.uptime ?? 0;
          const color = uptimeColor(uptime, hasData);
          const label = hasData
            ? `${formatDay(day)}: ${(uptime * 100).toFixed(1)}% uptime`
            : `${formatDay(day)}: no data`;
          const heightPct = hasData ? Math.max(20, Math.round(uptime * 100)) : 20;

          return (
            <div
              key={day}
              title={label}
              className={`flex-1 rounded-sm ${color}`}
              style={{ height: `${heightPct}%` }}
              aria-label={label}
            />
          );
        })}
      </div>

      {/* Axis labels */}
      <div className="mt-1 flex justify-between text-xs text-[var(--ink-500)]">
        <span>{formatDay(slots[0])}</span>
        <span>{formatDay(slots[slots.length - 1])}</span>
      </div>
    </div>
  );
}
