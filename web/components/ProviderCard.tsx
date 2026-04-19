import Link from "next/link";
import type { ProviderSummary, ModelStat } from "@/lib/api";
import { LatencySparkline } from "./LatencySparkline";

// ── Status pill ────────────────────────────────────────────────────────────

const STATUS_STYLES = {
  operational: "bg-[var(--signal-ok-bg)] text-[var(--signal-ok)]",
  degraded:    "bg-[var(--signal-warn-bg)] text-[var(--signal-warn)]",
  down:        "bg-[var(--signal-down-bg)] text-[var(--signal-down)]",
} as const;

function StatusPill({ status }: { status: string }) {
  const cls = STATUS_STYLES[status as keyof typeof STATUS_STYLES]
    ?? "bg-[var(--canvas-overlay)] text-[var(--ink-400)]";
  return (
    <span className={`rounded px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide ${cls}`}>
      {status}
    </span>
  );
}

// ── Stat cell ──────────────────────────────────────────────────────────────

function uptimeColor(u: number): string {
  if (u >= 0.995) return "text-[var(--signal-ok)]";
  if (u >= 0.95)  return "text-[var(--signal-warn)]";
  return "text-[var(--signal-down)]";
}

function StatCell({ label, value, colorClass }: { label: string; value: string; colorClass?: string }) {
  return (
    <div className="flex flex-col gap-0.5">
      <span className="text-[9px] font-semibold uppercase tracking-[0.1em] text-[var(--ink-500)]">
        {label}
      </span>
      <span className={`text-sm font-mono tabular-nums ${colorClass ?? "text-[var(--ink-300)]"}`}>
        {value}
      </span>
    </div>
  );
}

// ── Provider-level sparkline ───────────────────────────────────────────────
// Average non-zero avg_ms values across all models per time bucket.

function aggregateSparklines(modelStats: ModelStat[]): number[] {
  const BUCKETS = 60;
  const sums = new Array<number>(BUCKETS).fill(0);
  const counts = new Array<number>(BUCKETS).fill(0);
  for (const m of modelStats) {
    for (let i = 0; i < BUCKETS; i++) {
      if (m.sparkline[i] > 0) {
        sums[i] += m.sparkline[i];
        counts[i]++;
      }
    }
  }
  return sums.map((s, i) => (counts[i] > 0 ? s / counts[i] : 0));
}

// ── Provider card ──────────────────────────────────────────────────────────

interface Props {
  provider: ProviderSummary;
}

export function ProviderCard({ provider: p }: Props) {
  const uptime = p.uptime_24h !== undefined
    ? (p.uptime_24h * 100).toFixed(1) + "%"
    : "—";

  const p95 = p.p95_ms !== undefined
    ? p.p95_ms < 1000
      ? `${Math.round(p.p95_ms)}ms`
      : `${(p.p95_ms / 1000).toFixed(1)}s`
    : "—";

  const modelNames = (p.model_stats ?? []).map((m) => m.display_name);
  const sparkline = aggregateSparklines(p.model_stats ?? []);

  return (
    <Link
      href={`/providers/${p.id}`}
      className="flex flex-col rounded-lg border border-[var(--ink-600)] bg-[var(--canvas-raised)]
                 hover:border-[var(--ink-500)] hover:bg-[var(--canvas-overlay)]
                 transition-colors"
    >
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3">
        <span className="text-sm font-semibold text-[var(--ink-100)]">{p.name}</span>
        <StatusPill status={p.current_status} />
      </div>

      {/* Stats + sparkline */}
      <div className="flex items-end gap-4 border-t border-[var(--ink-600)] px-4 py-3">
        <StatCell
          label="Uptime 24h"
          value={uptime}
          colorClass={p.uptime_24h !== undefined ? uptimeColor(p.uptime_24h) : "text-[var(--ink-500)]"}
        />
        <StatCell label="p95" value={p95} />
        <div className="flex-1">
          <LatencySparkline data={sparkline} width={200} height={28} className="w-full" />
        </div>
      </div>

      {/* Model tags */}
      {modelNames.length > 0 && (
        <div className="flex flex-wrap gap-1 border-t border-[var(--ink-600)] px-4 pb-3 pt-2">
          {modelNames.slice(0, 4).map((name) => (
            <span
              key={name}
              className="rounded bg-[var(--canvas-overlay)] px-1.5 py-0.5 text-[10px] text-[var(--ink-500)]"
            >
              {name}
            </span>
          ))}
          {modelNames.length > 4 && (
            <span className="px-1 py-0.5 text-[10px] text-[var(--ink-600)]">
              +{modelNames.length - 4} more
            </span>
          )}
        </div>
      )}
    </Link>
  );
}
