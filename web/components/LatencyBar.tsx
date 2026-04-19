import type { HistoryBucket, HistoryWindow } from "@/lib/api";

interface Props {
  buckets: HistoryBucket[];
  window?: HistoryWindow;
  /** Legacy: number of daily slots; ignored when `window` is set. */
  days?: number;
}

type Config = { count: number; keyLen: number; label: string; granularity: "hour" | "day" };

function windowConfig(w: HistoryWindow | undefined, days: number): Config {
  switch (w) {
    case "24h": return { count: 24, keyLen: 13, label: "p95 latency — 24h",     granularity: "hour" };
    case "7d":  return { count: 7,  keyLen: 10, label: "p95 latency — 7 days",  granularity: "day"  };
    default:    return { count: days, keyLen: 10, label: `p95 latency — ${days} days`, granularity: "day" };
  }
}

function makeSlots(cfg: Config): string[] {
  const slots: string[] = [];
  const now = new Date();
  for (let i = cfg.count - 1; i >= 0; i--) {
    const d = new Date(now);
    if (cfg.granularity === "hour") {
      d.setUTCHours(d.getUTCHours() - i, 0, 0, 0);
    } else {
      d.setUTCDate(d.getUTCDate() - i);
    }
    slots.push(d.toISOString().slice(0, cfg.keyLen));
  }
  return slots;
}

function formatSlot(key: string, granularity: "hour" | "day"): string {
  if (granularity === "hour") {
    return `${key.slice(11, 13)}:00`;
  }
  return new Date(key + "T00:00:00Z").toLocaleDateString("en-US", {
    month: "short", day: "numeric", timeZone: "UTC",
  });
}

const CEILING_MS = 3000;

function latencyColor(p95Ms: number, hasData: boolean): string {
  if (!hasData) return "bg-[var(--ink-600)]";
  if (p95Ms <= 500)  return "bg-[var(--signal-ok)]";
  if (p95Ms <= 2000) return "bg-[var(--signal-warn)]";
  return "bg-[var(--signal-down)]";
}

function formatMs(ms: number): string {
  return ms >= 1000 ? `${(ms / 1000).toFixed(2)}s` : `${Math.round(ms)}ms`;
}

export function LatencyBar({ buckets, window, days = 30 }: Props) {
  const cfg = windowConfig(window, days);
  const slots = makeSlots(cfg);

  const bySlot = new Map<string, HistoryBucket>();
  for (const b of buckets) {
    bySlot.set(b.timestamp.slice(0, cfg.keyLen), b);
  }

  const withData = buckets.filter((b) => b.total > 0 && b.p95_ms > 0);
  const medianP95 =
    withData.length === 0
      ? null
      : withData.map((b) => b.p95_ms).sort((a, b) => a - b)[Math.floor(withData.length / 2)];

  return (
    <div>
      <div className="mb-2 flex items-center justify-between text-xs text-[var(--ink-400)]">
        <span>{cfg.label}</span>
        {medianP95 !== null && (
          <span className="font-medium text-[var(--ink-200)]">
            {formatMs(medianP95)} median
          </span>
        )}
      </div>

      <div className="flex h-10 items-end gap-px" aria-label={`${cfg.label} chart`}>
        {slots.map((slot) => {
          const b = bySlot.get(slot);
          const hasData = b !== undefined && b.total > 0 && b.p95_ms > 0;
          const p95 = b?.p95_ms ?? 0;
          const color = latencyColor(p95, hasData);
          const label = hasData
            ? `${formatSlot(slot, cfg.granularity)}: ${formatMs(p95)} p95`
            : `${formatSlot(slot, cfg.granularity)}: no data`;
          const heightPct = hasData
            ? Math.max(10, Math.min(100, Math.round((p95 / CEILING_MS) * 100)))
            : 20;

          return (
            <div
              key={slot}
              title={label}
              className={`flex-1 rounded-sm ${color}`}
              style={{ height: `${heightPct}%` }}
              aria-label={label}
            />
          );
        })}
      </div>

      <div className="mt-1 flex justify-between text-xs text-[var(--ink-500)]">
        <span>{formatSlot(slots[0], cfg.granularity)}</span>
        <span>{formatSlot(slots[slots.length - 1], cfg.granularity)}</span>
      </div>
    </div>
  );
}
