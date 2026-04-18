import type { HistoryBucket } from "@/lib/api";

interface Props {
  buckets: HistoryBucket[];
  days?: number;
}

// Reference ceiling for bar scaling. Bars at or above this value fill 100% height.
const CEILING_MS = 3000;

function latencyColor(p95Ms: number, hasData: boolean): string {
  if (!hasData) return "bg-[var(--ink-600)]";
  if (p95Ms <= 500) return "bg-[var(--signal-ok)]";
  if (p95Ms <= 2000) return "bg-[var(--signal-warn)]";
  return "bg-[var(--signal-down)]";
}

function formatMs(ms: number): string {
  if (ms >= 1000) return `${(ms / 1000).toFixed(2)}s`;
  return `${Math.round(ms)}ms`;
}

function formatDay(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    timeZone: "UTC",
  });
}

export function LatencyBar({ buckets, days = 30 }: Props) {
  const byDay = new Map<string, HistoryBucket>();
  for (const b of buckets) {
    byDay.set(b.timestamp.slice(0, 10), b);
  }

  const slots: string[] = [];
  const now = new Date();
  for (let i = days - 1; i >= 0; i--) {
    const d = new Date(now);
    d.setUTCDate(d.getUTCDate() - i);
    slots.push(d.toISOString().slice(0, 10));
  }

  // Summary: median p95 across all buckets that have data.
  const withData = buckets.filter((b) => b.total > 0 && b.p95_ms > 0);
  const medianP95 =
    withData.length === 0
      ? null
      : withData
          .map((b) => b.p95_ms)
          .sort((a, b) => a - b)[Math.floor(withData.length / 2)];

  return (
    <div>
      <div className="mb-2 flex items-center justify-between text-xs text-[var(--ink-400)]">
        <span>p95 latency — 30 days</span>
        {medianP95 !== null && (
          <span className="font-medium text-[var(--ink-200)]">
            {formatMs(medianP95)} median
          </span>
        )}
      </div>

      <div className="flex h-10 items-end gap-px" aria-label="30-day p95 latency chart">
        {slots.map((day) => {
          const b = byDay.get(day);
          const hasData = b !== undefined && b.total > 0 && b.p95_ms > 0;
          const p95 = b?.p95_ms ?? 0;
          const color = latencyColor(p95, hasData);
          const label = hasData
            ? `${formatDay(day)}: ${formatMs(p95)} p95`
            : `${formatDay(day)}: no data`;
          const heightPct = hasData
            ? Math.max(10, Math.min(100, Math.round((p95 / CEILING_MS) * 100)))
            : 20;

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

      <div className="mt-1 flex justify-between text-xs text-[var(--ink-500)]">
        <span>{formatDay(slots[0])}</span>
        <span>{formatDay(slots[slots.length - 1])}</span>
      </div>
    </div>
  );
}
