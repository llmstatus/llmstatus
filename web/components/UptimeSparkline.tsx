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
    case "24h": return { count: 24, keyLen: 13, label: "24h uptime",    granularity: "hour" };
    case "7d":  return { count: 7,  keyLen: 10, label: "7-day uptime",  granularity: "day"  };
    default:    return { count: days, keyLen: 10, label: `${days}-day uptime`, granularity: "day" };
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
    // key like "2026-04-19T14" → "14:00"
    const hour = key.slice(11, 13);
    return `${hour}:00`;
  }
  return new Date(key + "T00:00:00Z").toLocaleDateString("en-US", {
    month: "short", day: "numeric", timeZone: "UTC",
  });
}

function uptimeColor(uptime: number, hasData: boolean): string {
  if (!hasData) return "bg-[var(--ink-600)]";
  if (uptime >= 0.99) return "bg-[var(--signal-ok)]";
  if (uptime >= 0.95) return "bg-[var(--signal-warn)]";
  return "bg-[var(--signal-down)]";
}

export function UptimeSparkline({ buckets, window, days = 30 }: Props) {
  const cfg = windowConfig(window, days);
  const slots = makeSlots(cfg);

  const bySlot = new Map<string, HistoryBucket>();
  for (const b of buckets) {
    bySlot.set(b.timestamp.slice(0, cfg.keyLen), b);
  }

  const overall =
    buckets.length === 0
      ? null
      : buckets.reduce((sum, b) => sum + b.uptime, 0) / buckets.length;

  return (
    <div>
      <div className="mb-2 flex items-center justify-between text-xs text-[var(--ink-400)]">
        <span>{cfg.label}</span>
        {overall !== null && (
          <span className="font-medium text-[var(--ink-200)]">
            {(overall * 100).toFixed(2)} %
          </span>
        )}
      </div>

      <div className="flex h-10 items-end gap-px" aria-label={`${cfg.label} chart`}>
        {slots.map((slot) => {
          const b = bySlot.get(slot);
          const hasData = b !== undefined && b.total > 0;
          const uptime = b?.uptime ?? 0;
          const color = uptimeColor(uptime, hasData);
          const label = hasData
            ? `${formatSlot(slot, cfg.granularity)}: ${(uptime * 100).toFixed(1)}% uptime`
            : `${formatSlot(slot, cfg.granularity)}: no data`;
          const heightPct = hasData ? Math.max(20, Math.round(uptime * 100)) : 20;

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
